package jeu

import "fmt"

// Combats. Deux pions de la même pièce peuvent se battre : à mains nues,
// à la matraque ou au couteau sur des cases adjacentes ; au revolver ou au
// couteau lancé à distance, chaque case qui sépare les adversaires ôtant un
// point au dé. Les pare-balles ôtent 1 ou 2 points au dé de l'adversaire.
// Le revolver s'entend dans toute la villa (on peut s'en servir comme
// matraque, sans bruit).

// armesContre : les armes du voleur utilisables à cette distance.
func armesContre(q Equipement, dist int) []Arme {
	if dist <= 1 {
		l := []Arme{AMainNue}
		if q.Matraque || q.Revolver {
			l = append(l, AMatraque) // la crosse du revolver fait matraque
		}
		if q.Couteau {
			l = append(l, ACouteau)
		}
		if q.Revolver {
			l = append(l, ARevolver)
		}
		return l
	}
	var l []Arme
	if q.Couteau {
		l = append(l, ACouteau)
	}
	if q.Revolver {
		l = append(l, ARevolver)
	}
	return l
}

// Adversaires : qui le voleur peut attaquer d'ici.
func (e *Etat) Adversaires(v *Voleur) []Adversaire {
	var l []Adversaire
	k := v.Case
	if k < 0 || P.Terrain[k] == TEnceinte || P.Terrain[k] == TGrille {
		return nil
	}
	ajout := func(t TypeAdv, i, c int) {
		d := Dist(k, c)
		if a := armesContre(v.Equip, d); len(a) > 0 && P.MemeRegion(k, c) {
			l = append(l, Adversaire{Type: t, Indice: i, Case: c, Dist: d, Armes: a})
		}
	}
	for _, g := range e.Gardes {
		if g.Actif(e.Tour) {
			ajout(AdvGarde, g.Num, g.Case)
		}
	}
	for _, w := range e.Voleurs {
		if w != v && w.SurPlateau() && P.Terrain[w.Case] != TEnceinte && P.Terrain[w.Case] != TGrille {
			ajout(AdvVoleur, w.Num, w.Case)
		}
	}
	for _, m := range e.Meubles {
		if m.Type == MLit && m.Etat != Cache && (m.Dormeur == Endormi || m.Dormeur == Eveille) {
			ajout(AdvDormeur, m.Num, m.Tete())
		}
	}
	return l
}

func (e *Etat) phaseCombat(v *Voleur, ctrl Controleur) {
	o := &Options{Phase: PhaseCombat, Adversaires: e.Adversaires(v)}
	if len(o.Adversaires) == 0 {
		return
	}
	ch := ctrl.Agir(e, v, o)
	if ch.Type != CAttaquer || ch.Indice < 0 || ch.Indice >= len(o.Adversaires) {
		return
	}
	a := o.Adversaires[ch.Indice]
	if !contientArme(a.Armes, ch.Arme) {
		ch.Arme = a.Armes[0]
	}
	switch a.Type {
	case AdvGarde:
		e.combatGarde(v, e.Gardes[a.Indice], ch.Arme)
	case AdvVoleur:
		e.combatVoleurs(v, e.Voleurs[a.Indice], ch.Arme)
	case AdvDormeur:
		e.combatDormeur(v, e.Meubles[a.Indice], ch.Arme)
	}
	e.verifierVue(v)
}

func contientArme(l []Arme, a Arme) bool {
	for _, b := range l {
		if a == b {
			return true
		}
	}
	return false
}

// coup : un jet de combat ; retrait de la distance (armes de jet) et des
// pare-balles du défenseur.
func (e *Etat) coup(nomAtt, nomDef string, j int, a Arme, dist, pareBalles int) Issue {
	d := e.De()
	mod := 0
	if a == ACouteau || a == ARevolver {
		mod -= max(dist-1, 0)
	}
	mod -= pareBalles
	r := ResultatCombat(a, d+mod)
	txt := fmt.Sprintf("%s attaque %s %s", nomAtt, Le(nomDef), NomsArmes[a])
	if mod != 0 {
		txt += fmt.Sprintf(" (%+d)", mod)
	}
	e.evenement(Ev{Type: EvDes, Joueur: j, Titre: "Combat", Des: []int{d}, Texte: txt + " : " + NomsIssues[r]})
	e.evenement(Ev{Type: EvCombat, Joueur: j, Texte: txt + " : " + NomsIssues[r] + "."})
	return r
}

// coupDeFeu : le revolver donne l'alarme dans toute la villa ; tous les
// dormeurs se réveillent.
func (e *Etat) coupDeFeu(src int) {
	e.dire(-1, "Un coup de feu claque dans la villa !")
	for _, m := range e.Meubles {
		if m.Type == MLit && m.Dormeur == Endormi {
			m.Dormeur = Eveille
			e.voirParDormeur(m)
		}
	}
	e.alarme(src)
}

// combatGarde : c'est toujours le voleur qui attaque ; le garde riposte au
// revolver tant que personne n'est hors de combat.
func (e *Etat) combatGarde(v *Voleur, g *Garde, a Arme) {
	v.Vu = true
	for range 12 {
		dist := Dist(v.Case, g.Case)
		r := e.coup(v.Nom, g.Nom, v.Num, a, dist, 0)
		if a == ARevolver {
			e.coupDeFeu(v.Case)
		}
		if r != OK {
			if r == KO {
				g.KO = e.Tour + 2
				v.AAssomme = true
				e.dire(v.Num, "%s s'effondre, assommé.", Majuscule(g.Nom))
			} else {
				g.Mort = true
				v.ATue = true
				e.Cadavres = append(e.Cadavres, &Cadavre{Case: g.Case, Libelle: "le corps " + De(g.Nom)})
				e.dire(v.Num, "%s est tué.", Majuscule(g.Nom))
			}
			return
		}
		// riposte du garde
		r = e.coup(Majuscule(g.Nom), v.Nom, v.Num, ARevolver, dist, v.Equip.PareBalles)
		v.Vu = true
		e.coupDeFeu(g.Case)
		if r != OK {
			e.blesserVoleur(v, r, nil)
			return
		}
		if dist > 1 && a != ACouteau && a != ARevolver {
			return
		}
	}
}

// blesserVoleur : KO (deux tours sur place) ou mort ; le vainqueur (un
// voleur) peut s'emparer du butin.
func (e *Etat) blesserVoleur(v *Voleur, r Issue, vainqueur *Voleur) {
	if r == KO {
		v.Immobile = 2
		e.dire(v.Num, "%s est assommé : 2 tours sur place.", v.Nom)
	} else {
		v.Statut, v.TourFin = Abattu, e.Tour
		e.Cadavres = append(e.Cadavres, &Cadavre{Case: v.Case, Libelle: "le corps " + De(v.Nom)})
		e.dire(v.Num, "%s est tué.", v.Nom)
		v.Case = -1
	}
	if vainqueur != nil && len(v.Butins) > 0 {
		vainqueur.Butins = append(vainqueur.Butins, v.Butins...)
		e.dire(vainqueur.Num, "%s s'empare du butin %s (%s).", vainqueur.Nom, De(v.Nom), Francs(v.Valeur()))
		v.Butins = nil
	}
}

// meilleureArme : la riposte d'un voleur attaqué (couteau, matraque,
// mains nues ; le revolver en dernier recours à cause du bruit).
func meilleureArme(q Equipement, dist int) (Arme, bool) {
	l := armesContre(q, dist)
	for _, a := range []Arme{ACouteau, AMatraque, AMainNue, ARevolver} {
		if contientArme(l, a) {
			return a, true
		}
	}
	return 0, false
}

// combatVoleurs : chacun frappe à son tour jusqu'à ce que l'un soit hors
// de combat (ou que le défenseur ne puisse pas riposter).
func (e *Etat) combatVoleurs(att, def *Voleur, a Arme) {
	for range 12 {
		dist := Dist(att.Case, def.Case)
		r := e.coup(att.Nom, def.Nom, att.Num, a, dist, def.Equip.PareBalles)
		if a == ARevolver {
			e.coupDeFeu(att.Case)
		}
		if r != OK {
			e.blesserVoleur(def, r, att)
			return
		}
		ra, ok := meilleureArme(def.Equip, dist)
		if !ok {
			return
		}
		r = e.coup(def.Nom, att.Nom, def.Num, ra, dist, att.Equip.PareBalles)
		if ra == ARevolver {
			e.coupDeFeu(def.Case)
		}
		if r != OK {
			e.blesserVoleur(att, r, def)
			return
		}
	}
}

// combatDormeur : le voleur doit frapper avant que le dormeur donne
// l'alarme ; s'il le manque, le dormeur ne se bat pas mais appelle.
func (e *Etat) combatDormeur(v *Voleur, m *Meuble, a Arme) {
	dist := Dist(v.Case, m.Tete())
	r := e.coup(v.Nom, "le dormeur", v.Num, a, dist, 0)
	if a == ARevolver {
		e.coupDeFeu(v.Case)
	}
	switch r {
	case OK:
		v.Vu = true
		m.Dormeur = Eveille
		e.dire(v.Num, "Le dormeur se dresse et donne l'alarme !")
		e.alarme(m.Tete())
	case KO:
		v.AAssomme = true
		m.Dormeur, m.DormeurKO = Assomme, e.Tour+2
		e.dire(v.Num, "Le dormeur est assommé.")
	case Tue:
		v.ATue = true
		m.Dormeur = Mort
		e.Cadavres = append(e.Cadavres, &Cadavre{Case: m.Tete(), Libelle: "un corps dans un lit"})
		e.dire(v.Num, "Le dormeur ne se réveillera plus.")
	}
}
