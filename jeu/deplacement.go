package jeu

// Déplacement des voleurs. Un état de recherche est une case et, sur une
// case-seuil, le côté d'où l'on vient : on repart librement de ce côté,
// mais passer de l'autre côté demande d'avoir forcé l'obstacle ce tour-ci
// (depuis cette case, au début du tour).

type etatPas struct {
	k, cote int
}

// pas : le voleur peut-il aller de a (venu du côté cote) à la case
// voisine b ? Renvoie le coût, le nouveau côté, si le déplacement doit
// s'arrêter là (fin) et s'il se termine par un saut.
func (e *Etat) pas(v *Voleur, a, cote, b int) (cout, nCote int, fin, saut, ok bool) {
	d := Direction(a, b)
	if d < 0 || !P.Praticable(b) {
		return
	}
	ta, tb := P.Terrain[a], P.Terrain[b]
	// depuis le mur d'enceinte, on saute (voir Sauts) ; seule la grille
	// forcée laisse entrer à pied
	if ta == TEnceinte || ta == TGrille {
		if !(ta == TGrille && v.Ouverts[P.ObsCase[a]] && (tb == TJardin || tb == TArbre)) {
			return
		}
	}
	if (tb == TEnceinte || tb == TGrille) && ta != TJardin && ta != TArbre {
		return
	}
	if P.Liens[a][d] == LMur {
		return
	}
	if e.MeubleSur(b) != nil || e.GardeSur(b) != nil {
		return
	}
	cout, nCote = 1, RHors
	var obs *Obstacle
	de := P.Region(a)
	if o := P.ObsLien[a][d]; o >= 0 {
		if !v.Ouverts[o] {
			return
		}
		obs = &P.Obstacles[o]
	}
	if ta == TSeuil {
		de = cote
		vers := P.Region(b)
		if vers != RSeuil && vers != cote {
			o := P.ObsCase[a]
			if !v.Ouverts[o] {
				return
			}
			obs = &P.Obstacles[o]
		}
	}
	switch {
	case tb == TSeuil && ta == TSeuil:
		nCote = cote
	case tb == TSeuil:
		nCote = P.Region(a)
	}
	if tb == TEnceinte || tb == TGrille {
		fin = true // on grimpe sur le mur : on s'arrête
	}
	if obs != nil && obs.Type == OFenetre && !obs.Arbre {
		vers := P.Region(b)
		switch {
		case de == RDehors && vers >= 0:
			cout = 2 // escalade jusqu'à la fenêtre
		case de >= 0 && vers == RDehors:
			fin, saut = true, true
		}
	}
	ok = true
	return
}

// Trajets : les cases où le voleur peut finir son déplacement avec pm
// points (plus court chemin en points, puis en cases).
func (e *Etat) Trajets(v *Voleur, pm int) map[int]Trajet {
	type noeud struct {
		cout, pas int
		prec      etatPas
		fin, saut bool
	}
	depart := etatPas{v.Case, v.Cote}
	vu := map[etatPas]noeud{depart: {prec: etatPas{-1, 0}}}
	file := []etatPas{depart}
	for len(file) > 0 {
		// file triée par coût (les coûts sont petits)
		bi := 0
		for i, s := range file {
			if vu[s].cout < vu[file[bi]].cout {
				bi = i
			}
		}
		s := file[bi]
		file = append(file[:bi], file[bi+1:]...)
		n := vu[s]
		if n.fin {
			continue
		}
		for d := range 6 {
			b := Voisin(s.k, d)
			if b < 0 {
				continue
			}
			cout, cote, fin, saut, ok := e.pas(v, s.k, s.cote, b)
			if !ok || n.cout+cout > pm {
				continue
			}
			t := etatPas{b, cote}
			if old, deja := vu[t]; deja && (old.cout < n.cout+cout || (old.cout == n.cout+cout && old.pas <= n.pas+1)) {
				continue
			}
			vu[t] = noeud{cout: n.cout + cout, pas: n.pas + 1, prec: s, fin: fin, saut: saut}
			file = append(file, t)
		}
	}
	res := map[int]Trajet{}
	for s, n := range vu {
		if s == depart || e.VoleurSur(s.k) != nil {
			continue
		}
		if old, deja := res[s.k]; deja && old.Cout <= n.cout {
			continue
		}
		var ch []int
		for c := s; c != depart; c = vu[c].prec {
			ch = append([]int{c.k}, ch...)
		}
		res[s.k] = Trajet{Chemin: ch, Cout: n.cout, Saut: n.saut}
	}
	return res
}

// Sauts : depuis une case-mur, les cases du jardin où l'on peut sauter.
func (e *Etat) Sauts(v *Voleur) []int {
	if v.Case < 0 || (P.Terrain[v.Case] != TEnceinte && P.Terrain[v.Case] != TGrille) {
		return nil
	}
	var l []int
	for d := range 6 {
		b := Voisin(v.Case, d)
		if b < 0 || (P.Terrain[b] != TJardin && P.Terrain[b] != TArbre) || P.Liens[v.Case][d] == LMur {
			continue
		}
		if e.MeubleSur(b) == nil && e.GardeSur(b) == nil && e.VoleurSur(b) == nil {
			l = append(l, b)
		}
	}
	return l
}

// ---------------------------------------------------------------- gardes

// passeGarde : un garde peut-il aller de a à b ? En ronde, il reste dans
// la villa et ne franchit que les portes intérieures ; à la poursuite d'un
// voleur ou vers une alarme, il passe aussi les portes extérieures et va
// dans le jardin (jamais par les fenêtres ni sur le mur).
func (e *Etat) passeGarde(a, b int, ronde bool) bool {
	d := Direction(a, b)
	if d < 0 || P.Liens[a][d] == LMur || e.MeubleSur(b) != nil {
		return false
	}
	switch P.Terrain[b] {
	case TPiece:
	case TSeuil:
		o := P.Obstacles[P.ObsCase[b]]
		if o.Type == OFenetre || (ronde && o.Type == OPorteExt) {
			return false
		}
	case TJardin, TArbre:
		if ronde {
			return false
		}
	default:
		return false
	}
	if P.Terrain[a] == TSeuil {
		o := P.Obstacles[P.ObsCase[a]]
		if o.Type == OFenetre {
			return false
		}
	}
	if o := P.ObsLien[a][d]; o >= 0 {
		t := P.Obstacles[o].Type
		if t == OFenetre || (ronde && t == OPorteExt) {
			return false
		}
	}
	return true
}

// cheminGarde : premier pas du plus court chemin de g vers la cible (ou
// une case voisine si la cible est inaccessible), -1 s'il n'y en a pas.
func (e *Etat) cheminGarde(g *Garde, cible int) (int, int) {
	prec := map[int]int{g.Case: -1}
	file := []int{g.Case}
	best, bd := -1, 1<<30
	for len(file) > 0 {
		k := file[0]
		file = file[1:]
		if dd := Dist(k, cible); dd < bd {
			best, bd = k, dd
		}
		if k == cible {
			break
		}
		for d := range 6 {
			b := Voisin(k, d)
			if b < 0 {
				continue
			}
			if _, ok := prec[b]; ok {
				continue
			}
			// un voleur ne bloque pas : on va le prendre ; l'autre garde, si
			if b != cible && e.GardeSur(b) != nil {
				continue
			}
			if !e.passeGarde(k, b, false) {
				continue
			}
			prec[b] = k
			file = append(file, b)
		}
	}
	if best == g.Case || best < 0 {
		return -1, bd
	}
	k := best
	for prec[k] != g.Case {
		k = prec[k]
	}
	return k, bd
}
