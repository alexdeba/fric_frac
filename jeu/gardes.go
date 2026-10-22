package jeu

// Les gardes. En ronde, chacun va en ligne droite dans la direction tirée
// au dé (les chiffres autour de sa base) jusqu'à un obstacle (mur, meuble,
// fenêtre, porte extérieure), puis on relance le dé. Une case par joueur et
// par tour ; deux dès que l'alarme a été donnée ou qu'un voleur a été vu.

// deplacerGardes : à la fin du tour de chaque joueur.
func (e *Etat) deplacerGardes() {
	if !e.Ronde || e.Fini {
		return
	}
	n := 1
	if e.Alarme {
		n = 2
	}
	for range n {
		for _, g := range e.Gardes {
			e.pasGarde(g)
		}
	}
}

// percevoir : voleurs dans la même pièce que le garde, cadavres.
func (e *Etat) percevoir(g *Garde) {
	if !g.Actif(e.Tour) {
		return
	}
	var proie *Voleur
	for _, v := range e.Voleurs {
		if !v.SurPlateau() || !P.MemeRegion(g.Case, v.Case) {
			continue
		}
		if proie == nil || Dist(g.Case, v.Case) < Dist(g.Case, proie.Case) {
			proie = v
		}
	}
	if proie != nil {
		if g.Mode == Ronde || g.Proie != proie.Num { // une nouvelle découverte
			e.dire(-1, "%s aperçoit %s !", Majuscule(g.Nom), Le(proie.Nom))
		}
		proie.Vu = true
		g.Mode, g.Proie, g.Cible = Poursuite, proie.Num, proie.Case
		e.Alarme = true
	} else if g.Mode == Poursuite {
		// la proie est sortie de la pièce : on va où on l'a vue
		g.Mode = Alerte
	}
	for _, c := range e.Cadavres {
		if !c.Trouve && P.MemeRegion(g.Case, c.Case) {
			c.Trouve = true
			e.dire(-1, "%s découvre %s : l'alarme est donnée !", Majuscule(g.Nom), c.Libelle)
			e.alarme(c.Case)
		}
	}
}

func (e *Etat) pasGarde(g *Garde) {
	if !g.Actif(e.Tour) || e.Fini {
		return
	}
	e.percevoir(g)
	var b int
	switch g.Mode {
	case Poursuite, Alerte:
		if g.Mode == Poursuite {
			g.Cible = e.Voleurs[g.Proie].Case
		}
		if g.Case == g.Cible {
			g.Mode = Ronde
			g.Dir = g.Dirs[e.De()-1]
			b = e.pasRonde(g)
			break
		}
		var reste int
		b, reste = e.cheminGarde(g, g.Cible)
		if b < 0 || (reste > 0 && g.Mode == Alerte && Dist(g.Case, g.Cible) <= reste) {
			// on ne s'approche plus : arrivé au plus près de l'alarme
			g.Mode = Ronde
			g.Dir = g.Dirs[e.De()-1]
			b = e.pasRonde(g)
		}
	default:
		b = e.pasRonde(g)
	}
	if b < 0 || e.GardeSur(b) != nil {
		return
	}
	a := g.Case
	g.Case = b
	e.evenement(Ev{Type: EvGarde, Joueur: g.Num, Chemin: []int{a, b}})
	if v := e.VoleurSur(b); v != nil {
		e.arreter(v, g)
	}
	if g.Mode == Alerte && g.Case == g.Cible {
		g.Mode = Ronde
	}
	e.percevoir(g)
}

// pasRonde : la case suivante en ligne droite, ou une nouvelle direction
// tirée au dé tant que la voie est bloquée.
func (e *Etat) pasRonde(g *Garde) int {
	libre := func(d int) int {
		b := Voisin(g.Case, d)
		if b < 0 || !e.passeGarde(g.Case, b, true) || e.GardeSur(b) != nil {
			return -1
		}
		return b
	}
	if b := libre(g.Dir); b >= 0 {
		return b
	}
	ouvertes := 0
	for d := range 6 {
		if libre(d) >= 0 {
			ouvertes++
		}
	}
	if ouvertes == 0 {
		return -1
	}
	for {
		g.Dir = g.Dirs[e.De()-1]
		if b := libre(g.Dir); b >= 0 {
			return b
		}
	}
}

// arreter : le garde se place sur la case du voleur.
func (e *Etat) arreter(v *Voleur, g *Garde) {
	v.Statut, v.TourFin = Arrete, e.Tour
	e.evenement(Ev{Type: EvArrestation, Joueur: v.Num, Case: v.Case, Indice: g.Num})
	e.dire(v.Num, "%s met la main au collet %s. Arrêté !", Majuscule(g.Nom), De(v.Nom))
	v.Case = -1
	for _, h := range e.Gardes {
		if h.Mode == Poursuite && h.Proie == v.Num {
			h.Mode = Ronde
			h.Dir = h.Dirs[e.De()-1]
		}
	}
}
