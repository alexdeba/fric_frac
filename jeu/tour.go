package jeu

import "fmt"

// Jouer : la partie entière. À chaque tour, chaque voleur force, ramasse,
// se déplace, combat éventuellement, puis déplace les deux gardes.
func (e *Etat) Jouer() {
	for _, v := range e.Voleurs {
		q := e.Ctrl[v.Num].Equiper(e, v)
		if !q.Valide() {
			q = Equipement{}
		}
		v.Equip = q
		e.dire(v.Num, "%s s'équipe : %s.", v.Nom, q.Texte())
	}
	for e.Tour = 1; !e.Fini; e.Tour++ {
		if e.Tour > TourMax {
			e.leverDuJour()
			break
		}
		for _, v := range e.Voleurs {
			if !v.Actif() {
				continue
			}
			e.Courant = v.Num
			e.evenement(Ev{Type: EvTour, Joueur: v.Num})
			e.tourVoleur(v)
			e.reveilDormeurs()
			if e.compterActifs() == 0 {
				break
			}
		}
		if e.compterActifs() == 0 {
			break
		}
	}
	e.Fini = true
	e.evenement(Ev{Type: EvFin})
}

func (e *Etat) compterActifs() int {
	n := 0
	for _, v := range e.Voleurs {
		if v.Actif() {
			n++
		}
	}
	return n
}

func (e *Etat) leverDuJour() {
	for _, v := range e.Voleurs {
		if v.Actif() {
			v.Statut, v.TourFin = Arrete, e.Tour
			e.dire(v.Num, "Le jour se lève : %s est cueilli par la police.", v.Nom)
		}
	}
}

func (e *Etat) tourVoleur(v *Voleur) {
	v.Ouverts = map[int]bool{}
	ctrl := e.Ctrl[v.Num]
	defer e.deplacerGardes()
	if v.Immobile > 0 {
		v.Immobile--
		e.dire(v.Num, "%s reste sur place.", v.Nom)
		return
	}
	if t := v.Travail; t != nil {
		t.Restants--
		if t.Restants > 0 {
			e.dire(v.Num, "%s s'applique sur %s (encore %d tour%s).", v.Nom, NomsForcables[t.Eff.Forcable], t.Restants, pluriel(t.Restants))
			return
		}
		v.Travail = nil
		e.forcer(v, t.Eff, t.Tours)
		if !v.SurPlateau() || v.Immobile > 0 {
			return
		}
	}
	if v.Statut == Attente {
		o := &Options{Phase: PhaseEntree, Entrees: e.entreesLibres()}
		ch := ctrl.Agir(e, v, o)
		if ch.Type != CEntrer || !contient(o.Entrees, ch.Case) {
			ch.Case = o.Entrees[e.rng.IntN(len(o.Entrees))]
		}
		v.Statut, v.Case, v.Cote = EnJeu, ch.Case, RHors
		e.evenement(Ev{Type: EvMouvement, Joueur: v.Num, Chemin: []int{ch.Case}})
		if P.Terrain[ch.Case] == TGrille {
			e.dire(v.Num, "%s se poste devant la grille.", v.Nom)
		} else {
			e.dire(v.Num, "%s grimpe sur le mur d'enceinte.", v.Nom)
			return
		}
	}
	// effractions (avant tout déplacement)
	for v.SurPlateau() && v.Immobile == 0 {
		o := &Options{Phase: PhaseEffraction, Effractions: e.EffractionsPossibles(v)}
		if len(o.Effractions) == 0 {
			break
		}
		ch := ctrl.Agir(e, v, o)
		if ch.Type != CForcer || ch.Indice < 0 || ch.Indice >= len(o.Effractions) {
			break
		}
		eff := o.Effractions[ch.Indice]
		n := min(max(ch.Tours, 1), 8)
		if n > 1 {
			v.Travail = &Travail{Eff: eff, Tours: n, Restants: n - 1}
			e.dire(v.Num, "%s entreprend de forcer %s (%d tours).", v.Nom, NomsForcables[eff.Forcable], n)
			return
		}
		e.forcer(v, eff, 1)
	}
	if !v.SurPlateau() || v.Immobile > 0 {
		return
	}
	e.marche(v, ctrl)
	if !v.SurPlateau() || v.Immobile > 0 {
		return
	}
	e.phaseCombat(v, ctrl)
}

func pluriel(n int) string {
	if n > 1 {
		return "s"
	}
	return ""
}

func contient(l []int, x int) bool {
	for _, y := range l {
		if y == x {
			return true
		}
	}
	return false
}

func (e *Etat) entreesLibres() []int {
	var l []int
	for _, k := range P.Enceinte {
		if e.VoleurSur(k) == nil {
			l = append(l, k)
		}
	}
	return l
}

// ---------------------------------------------------------------- effraction

// EffractionsPossibles : obstacles de la case (case-seuil ou liaison
// touchant la case) pas encore forcés ce tour, meubles adjacents découverts
// par ce voleur.
func (e *Etat) EffractionsPossibles(v *Voleur) []Effraction {
	var l []Effraction
	k := v.Case
	ajout := func(f Forcable, o, m int) {
		eff := Effraction{Forcable: f, Obstacle: o, Meuble: m}
		for n := range 8 {
			eff.Bruits[n] = Bruit(f, n+1, v.Equip.Passe)
		}
		l = append(l, eff)
	}
	if o := P.ObsCase[k]; o >= 0 && !v.Ouverts[o] {
		ob := &P.Obstacles[o]
		// la grille ne se force que du dehors
		if ob.Type != OGrille || v.Cote == RHors {
			ajout(forcableObstacle(ob.Type), o, -1)
		}
	}
	for d := range 6 {
		if o := P.ObsLien[k][d]; o >= 0 && !v.Ouverts[o] {
			ajout(forcableObstacle(P.Obstacles[o].Type), o, -1)
		}
	}
	for _, m := range e.Meubles {
		if m.Etat == Decouvert && m.Decouvreur == v.Num && m.Adjacent(k) {
			if f, n := m.Type.Forcable(); n > 0 {
				ajout(f, -1, m.Num)
			}
		}
	}
	return l
}

func forcableObstacle(t TypeObstacle) Forcable {
	return [...]Forcable{FGrille, FPorteExt, FFenetre, FPorteInt}[t]
}

// forcer : l'effraction aboutit ; le bruit, puis le butin d'un meuble.
func (e *Etat) forcer(v *Voleur, eff Effraction, tours int) {
	r := Bruit(eff.Forcable, tours, v.Equip.Passe)
	src := v.Case
	if eff.Obstacle >= 0 {
		v.Ouverts[eff.Obstacle] = true
		ob := &P.Obstacles[eff.Obstacle]
		if ob.Case >= 0 {
			src = ob.Case
		}
	} else {
		src = e.Meubles[eff.Meuble].Cases[0]
	}
	if r > 0 {
		e.dire(v.Num, "%s force %s : bruit à %d case%s.", v.Nom, NomsForcables[eff.Forcable], r, pluriel(r))
	} else {
		e.dire(v.Num, "%s force %s sans un bruit.", v.Nom, NomsForcables[eff.Forcable])
	}
	e.bruit(src, r)
	if eff.Meuble >= 0 {
		m := e.Meubles[eff.Meuble]
		m.Etat = Force
		_, n := m.Type.Forcable()
		for i := range n {
			if !v.SurPlateau() {
				return
			}
			e.tirerButin(v, m, i+1, n)
		}
	}
}

func (e *Etat) tirerButin(v *Voleur, m *Meuble, i, n int) {
	d1, d2 := e.De(), e.De()
	b := TableButin(10*d1 + d2)
	titre := fmt.Sprintf("Butin (%d/%d)", i, n)
	e.evenement(Ev{Type: EvDes, Joueur: v.Num, Titre: titre, Des: []int{d1, d2}, Texte: b.Texte})
	switch b.Type {
	case BRien:
		e.dire(v.Num, "%s fouille : rien.", v.Nom)
	case BObjet:
		v.Butins = append(v.Butins, b.Objet)
		e.evenement(Ev{Type: EvButin, Joueur: v.Num, Texte: fmt.Sprintf("%s rafle %s (%s).", v.Nom, b.Objet.Nom, Francs(b.Objet.Valeur))})
	case BAlarme2, BAlarme3:
		nd := 2
		if b.Type == BAlarme3 {
			nd = 3
		}
		var des []int
		r := 0
		for range nd {
			x := e.De()
			des = append(des, x)
			r += x
		}
		e.evenement(Ev{Type: EvDes, Joueur: v.Num, Titre: "Signal d'alarme", Des: des})
		e.dire(v.Num, "Un signal d'alarme retentit : bruit à %d cases !", r)
		e.bruit(m.Cases[0], r)
	case BGaz:
		v.Immobile += 2
		e.dire(v.Num, "Gaz étourdissant : %s reste 2 tours sur place.", v.Nom)
	case BCharnieres:
		e.dire(v.Num, "Les charnières grincent : bruit à 3 cases.")
		e.bruit(m.Cases[0], 3)
	case BTombe:
		e.dire(v.Num, "Le meuble tombe : bruit à 4 cases.")
		e.bruit(m.Cases[0], 4)
	case BChalumeau:
		e.dire(v.Num, "Le chalumeau explose : bruit à 6 cases !")
		e.bruit(m.Cases[0], 6)
	case BGardes:
		e.dire(v.Num, "Les gardes avancent de 3 cases.")
		e.Ronde = true
		for range 3 {
			for _, g := range e.Gardes {
				e.pasGarde(g)
			}
		}
	case BPassage:
		v.Passage = m.Num
		e.dire(v.Num, "%s découvre un passage secret vers l'extérieur.", v.Nom)
	}
}

// ---------------------------------------------------------------- bruit

// bruit : un garde ou un dormeur à r cases ou moins l'entend, et l'alarme
// est donnée.
func (e *Etat) bruit(src, r int) {
	if r <= 0 {
		return
	}
	e.evenement(Ev{Type: EvBruit, Case: src, Rayon: r})
	entendu := false
	for _, g := range e.Gardes {
		if g.Actif(e.Tour) && Dist(g.Case, src) <= r {
			entendu = true
		}
	}
	for _, m := range e.Meubles {
		if m.Type == MLit && m.Dormeur == Endormi && Dist(m.Tete(), src) <= r {
			m.Dormeur = Eveille
			e.dire(-1, "Le dormeur %s se réveille !", Du(NomPiece(P.Region(m.Cases[0]))))
			e.voirParDormeur(m)
			entendu = true
		}
	}
	if entendu {
		e.alarme(src)
	}
}

// alarme : les gardes accourent vers la case suspecte, deux cases par tour
// jusqu'à la fin de la partie.
func (e *Etat) alarme(src int) {
	e.Ronde, e.Alarme = true, true
	for _, g := range e.Gardes {
		if g.Mort || g.Mode == Poursuite {
			continue
		}
		g.Mode, g.Cible = Alerte, src
	}
	e.evenement(Ev{Type: EvAlarme, Case: src, Texte: "L'alarme est donnée !"})
}

// voirParDormeur : un dormeur réveillé voit les voleurs de sa chambre.
func (e *Etat) voirParDormeur(m *Meuble) {
	for _, v := range e.Voleurs {
		if v.SurPlateau() && P.MemeRegion(v.Case, m.Tete()) {
			v.Vu = true
		}
	}
}

// reveilDormeurs : un dormeur assommé reprend ses esprits et donne l'alarme.
func (e *Etat) reveilDormeurs() {
	for _, m := range e.Meubles {
		if m.Type == MLit && m.Dormeur == Assomme && e.Tour >= m.DormeurKO {
			m.Dormeur = Eveille
			e.dire(-1, "Le dormeur %s reprend ses esprits et appelle à l'aide !", Du(NomPiece(P.Region(m.Cases[0]))))
			e.voirParDormeur(m)
			e.alarme(m.Tete())
		}
	}
}

// ---------------------------------------------------------------- marche

func (e *Etat) marche(v *Voleur, ctrl Controleur) {
	pm, pas := e.PM, 0
	for v.SurPlateau() && v.Immobile == 0 && pm > 0 {
		o := &Options{Phase: PhaseMarche, PM: pm, Trajets: e.Trajets(v, pm), Sauts: e.Sauts(v)}
		o.PeutSortir = P.Terrain[v.Case] == TEnceinte || P.Terrain[v.Case] == TGrille
		o.Passage = v.Passage >= 0 && e.Meubles[v.Passage].Adjacent(v.Case)
		if len(o.Trajets) == 0 && len(o.Sauts) == 0 && !o.PeutSortir && !o.Passage {
			break
		}
		ch := ctrl.Agir(e, v, o)
		switch ch.Type {
		case CAller:
			t, ok := o.Trajets[ch.Case]
			if !ok {
				ch.Type = Passer
				break
			}
			n, cout := e.suivre(v, t)
			pas += n
			pm -= cout
			if t.Saut && v.Case == t.Chemin[len(t.Chemin)-1] {
				e.sauter(v, "par la fenêtre")
				pm = 0
			}
			if v.Case >= 0 && (P.Terrain[v.Case] == TEnceinte || P.Terrain[v.Case] == TGrille) {
				pm = min(pm, 1) // sur le mur : il ne reste qu'à en descendre
			}
		case CSauter:
			if !contient(o.Sauts, ch.Case) {
				ch.Type = Passer
				break
			}
			e.evenement(Ev{Type: EvMouvement, Joueur: v.Num, Chemin: []int{v.Case, ch.Case}})
			v.Case, v.Cote = ch.Case, RHors
			e.sauter(v, "dans le jardin")
			e.apresPas(v)
			pm = 0
		case CSortir:
			if !o.PeutSortir {
				ch.Type = Passer
				break
			}
			e.sortir(v, "passe le mur et disparaît dans la nuit")
			return
		case CPassage:
			if !o.Passage {
				ch.Type = Passer
				break
			}
			e.sortir(v, "s'éclipse par le passage secret")
			return
		}
		if ch.Type == Passer {
			break
		}
	}
	if pas > 0 && v.SurPlateau() && !v.Equip.Chaussures {
		e.bruit(v.Case, pas)
	}
	e.verifierVue(v)
}

// suivre : le pion avance le long du trajet ; il s'arrête dès qu'il
// retourne un meuble (le joueur décide alors de continuer ou non).
func (e *Etat) suivre(v *Voleur, t Trajet) (int, int) {
	ch := []int{v.Case}
	pas, cout := 0, 0
	for _, b := range t.Chemin {
		c, cote, _, _, ok := e.pas(v, v.Case, v.Cote, b)
		if !ok {
			break
		}
		v.Case, v.Cote = b, cote
		ch = append(ch, b)
		pas++
		cout += c
		if e.apresPas(v) {
			break
		}
	}
	e.evenement(Ev{Type: EvMouvement, Joueur: v.Num, Chemin: ch})
	return pas, cout
}

// apresPas : meubles retournés (renvoie vrai si au moins un) et meubles
// abandonnés par leur découvreur ; la ronde commence dès qu'un voleur
// entre dans la villa.
func (e *Etat) apresPas(v *Voleur) bool {
	revele := false
	for _, m := range e.Meubles {
		switch {
		case m.Etat == Cache && m.Adjacent(v.Case):
			m.Etat, m.Decouvreur = Decouvert, v.Num
			revele = true
			e.evenement(Ev{Type: EvRevele, Joueur: v.Num, Indice: m.Num})
			e.dire(v.Num, "%s découvre %s.", v.Nom, NomsMeubles[m.Type])
			if m.Type == MLit {
				e.dire(v.Num, "Quelqu'un dort dans ce lit…")
			}
		case m.Etat == Decouvert && m.Decouvreur == v.Num && !m.Adjacent(v.Case):
			m.Etat = Vide
		}
	}
	if !e.Ronde && P.DansVilla(v.Case) {
		e.Ronde = true
		e.dire(-1, "Un voleur est entré dans la villa : les gardes commencent leur ronde.")
		for _, g := range e.Gardes {
			g.Dir = g.Dirs[e.De()-1]
		}
	}
	return revele
}

// sauter : du mur dans le jardin, ou d'une fenêtre sans arbre (table des
// combats, colonne saut ; les pare-balles ne protègent pas).
func (e *Etat) sauter(v *Voleur, ou string) {
	d := e.De()
	r := ResultatCombat(ASaut, d)
	e.evenement(Ev{Type: EvDes, Joueur: v.Num, Titre: "Saut", Des: []int{d}, Texte: NomsIssues[r]})
	if r == OK {
		e.dire(v.Num, "%s saute %s sans dommage.", v.Nom, ou)
		return
	}
	v.Immobile = 2
	e.dire(v.Num, "%s saute %s et se tord la cheville : 2 tours sur place.", v.Nom, ou)
}

func (e *Etat) sortir(v *Voleur, comment string) {
	v.Statut, v.TourFin = Sorti, e.Tour
	v.Score = v.Recel()
	e.evenement(Ev{Type: EvSortie, Joueur: v.Num, Case: v.Case})
	v.Case = -1
	e.dire(v.Num, "%s %s, avec %s de butin : %s chez le receleur.", v.Nom, comment, Francs(v.Valeur()), Francs(v.Score))
}

// verifierVue : un garde actif dans la même pièce voit le voleur et le
// prend en chasse.
func (e *Etat) verifierVue(v *Voleur) {
	if !v.SurPlateau() {
		return
	}
	for _, g := range e.Gardes {
		if g.Actif(e.Tour) && P.MemeRegion(g.Case, v.Case) && e.Ronde {
			if g.Mode != Poursuite {
				e.dire(-1, "%s aperçoit %s !", Majuscule(g.Nom), Le(v.Nom))
			}
			v.Vu = true
			g.Mode, g.Proie, g.Cible = Poursuite, v.Num, v.Case
			e.Alarme = true
		}
	}
	for _, m := range e.Meubles {
		if m.Type == MLit && m.Dormeur == Eveille && P.MemeRegion(m.Tete(), v.Case) {
			v.Vu = true
		}
	}
}

func Majuscule(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if r[0] >= 'a' && r[0] <= 'z' {
		r[0] -= 'a' - 'A'
	}
	return string(r)
}
