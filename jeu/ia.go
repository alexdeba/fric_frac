package jeu

import "math"

// IA : un voleur mené par l'ordinateur. Elle vise les meubles encore
// cachés (on voit les pions retournés face cachée, pas leur nature), force
// ceux qu'elle a découverts en réglant la durée de l'effraction pour que
// le bruit n'atteigne ni garde ni dormeur connu, et repart par le mur
// quand son butin dépasse son ambition ou que les gardes se rapprochent.
type IA struct{}

// ambition : le butin (valeur de la table) à partir duquel l'IA repart.
func ambition(v *Voleur) int {
	return []int{40000, 100000, 60000, 150000, 30000, 80000}[v.Num%6]
}

var kits = []Equipement{
	{Passe: 1, Chaussures: true},
	{Chaussures: true, Couteau: true},
	{Passe: 1, PareBalles: 1},
	{Chaussures: true, PareBalles: 1, Matraque: true},
	{Passe: 2},
	{Chaussures: true, Matraque: true, PareBalles: 1},
}

func (IA) Equiper(e *Etat, v *Voleur) Equipement {
	return kits[e.rng.IntN(len(kits))]
}

func (ia IA) Agir(e *Etat, v *Voleur, o *Options) Choix {
	switch o.Phase {
	case PhaseEntree:
		return ia.entrer(e, v, o)
	case PhaseEffraction:
		return ia.forcer(e, v, o)
	case PhaseMarche:
		return ia.marcher(e, v, o)
	case PhaseCombat:
		return ia.combattre(e, v, o)
	}
	return Choix{}
}

// partir : assez de butin, plus rien à chercher, ou trop de danger ; la
// décision est définitive.
func (IA) partir(e *Etat, v *Voleur) bool {
	if v.Repli {
		return true
	}
	if v.Valeur() >= ambition(v) || e.Tour > TourMax-25 {
		v.Repli = true
	}
	if v.Valeur() > 0 && e.Alarme && dangerGardes(e, v.Case) <= 5 {
		v.Repli = true
	}
	reste := false
	for _, m := range e.Meubles {
		if m.Etat == Cache || (m.Etat == Decouvert && m.Decouvreur == v.Num) {
			reste = true
		}
	}
	if !reste {
		v.Repli = true
	}
	return v.Repli
}

// dangerGardes : distance au garde actif le plus proche.
func dangerGardes(e *Etat, k int) int {
	d := 99
	for _, g := range e.Gardes {
		if g.Actif(e.Tour) && k >= 0 {
			d = min(d, Dist(g.Case, k))
		}
	}
	return d
}

// danger : distance au plus proche garde ou dormeur connu (lit retourné).
func danger(e *Etat, k int) int {
	d := dangerGardes(e, k)
	for _, m := range e.Meubles {
		if m.Type == MLit && m.Etat != Cache && m.Dormeur == Endormi {
			d = min(d, Dist(m.Tete(), k))
		}
	}
	return d
}

// ---------------------------------------------------------------- carte

// carte : coût estimé (en cases, les obstacles comptant pour un tour) de
// chaque case jusqu'au plus proche objectif.
func (ia IA) carte(e *Etat, v *Voleur) []int {
	const inf = math.MaxInt32
	dist := make([]int, NbCases)
	for i := range dist {
		dist[i] = inf
	}
	var file []int
	but := func(k int) {
		if k >= 0 && P.Praticable(k) && e.MeubleSur(k) == nil && dist[k] != 0 {
			dist[k] = 0
			file = append(file, k)
		}
	}
	if ia.partir(e, v) {
		for _, k := range P.Enceinte {
			if P.Terrain[k] == TEnceinte {
				but(k)
			}
		}
	} else {
		for _, m := range e.Meubles {
			if m.Etat == Cache || (m.Etat == Decouvert && m.Decouvreur == v.Num && lucratif(m)) {
				for _, c := range m.Cases {
					for d := range 6 {
						but(Voisin(c, d))
					}
				}
			}
		}
	}
	// Dijkstra à rebours (les coûts sont symétriques à peu près)
	for len(file) > 0 {
		bi := 0
		for i, k := range file {
			if dist[k] < dist[file[bi]] {
				bi = i
			}
		}
		k := file[bi]
		file = append(file[:bi], file[bi+1:]...)
		for d := range 6 {
			b := Voisin(k, d)
			if b < 0 || !P.Praticable(b) || P.Liens[k][d] == LMur || e.MeubleSur(b) != nil {
				continue
			}
			tb, tk := P.Terrain[b], P.Terrain[k]
			if (tb == TEnceinte || tb == TGrille) && tk != TJardin && tk != TArbre {
				continue
			}
			c := 1
			if P.ObsLien[k][d] >= 0 || tk == TSeuil {
				c += 3 // une effraction : un tour au moins
			}
			if tb == TEnceinte || tb == TGrille {
				c += 2 // le saut
			}
			if dist[k]+c < dist[b] {
				dist[b] = dist[k] + c
				file = append(file, b)
			}
		}
	}
	return dist
}

func lucratif(m *Meuble) bool { _, n := m.Type.Forcable(); return n > 0 }

// ---------------------------------------------------------------- phases

func (ia IA) entrer(e *Etat, v *Voleur, o *Options) Choix {
	// la case-mur la plus proche d'une porte ou d'une fenêtre de la villa
	best, bd := -1, 1<<30
	for _, k := range o.Entrees {
		if P.Terrain[k] != TEnceinte {
			continue
		}
		d := 99
		for _, ob := range P.Obstacles {
			if ob.Exterieur() && ob.Type != OGrille {
				d = min(d, int(math.Hypot(ob.Centre.X-Centre(k).X, ob.Centre.Y-Centre(k).Y)/GDX))
			}
		}
		d = d*2 + e.rng.IntN(10) // pas tous au même endroit
		if d < bd {
			best, bd = k, d
		}
	}
	return Choix{Type: CEntrer, Case: best}
}

// toursPour : le moins de tours possible pour que le bruit n'atteigne
// personne de connu.
func toursPour(e *Etat, f Forcable, passe, src int) int {
	dg := danger(e, src)
	for n := 1; n <= 8; n++ {
		if Bruit(f, n, passe) < dg {
			return n
		}
	}
	return ToursSilence(f, passe)
}

func (ia IA) forcer(e *Etat, v *Voleur, o *Options) Choix {
	if e.Alarme && dangerGardes(e, v.Case) <= 3 {
		return Choix{} // pas le temps : on file
	}
	carte := ia.carte(e, v)
	for i, eff := range o.Effractions {
		if eff.Meuble >= 0 {
			m := e.Meubles[eff.Meuble]
			return Choix{Type: CForcer, Indice: i, Tours: toursPour(e, eff.Forcable, v.Equip.Passe, m.Cases[0])}
		}
	}
	// un obstacle : seulement s'il ouvre le chemin vers l'objectif
	sans := ia.meilleur(e, v, carte, e.PM)
	for i, eff := range o.Effractions {
		v.Ouverts[eff.Obstacle] = true
		avec := ia.meilleur(e, v, carte, e.PM)
		delete(v.Ouverts, eff.Obstacle)
		if avec < sans-1 {
			src := v.Case
			if c := P.Obstacles[eff.Obstacle].Case; c >= 0 {
				src = c
			}
			return Choix{Type: CForcer, Indice: i, Tours: toursPour(e, eff.Forcable, v.Equip.Passe, src)}
		}
	}
	return Choix{}
}

// meilleur : la meilleure valeur de carte atteignable ce tour.
func (ia IA) meilleur(e *Etat, v *Voleur, carte []int, pm int) int {
	best := carte[v.Case]
	for k := range e.Trajets(v, pm) {
		best = min(best, carte[k])
	}
	return best
}

// risque : bruit de marche qui atteindrait un garde ou un dormeur connu,
// garde trop proche.
func risque(e *Etat, v *Voleur, k, pas int) int {
	r := 0
	if !v.Equip.Chaussures && pas > 0 && danger(e, k) <= pas {
		r += 40
	}
	switch dg := dangerGardes(e, k); {
	case dg <= 1:
		r += 60
	case dg <= 3:
		r += 15
	}
	return r
}

func (ia IA) marcher(e *Etat, v *Voleur, o *Options) Choix {
	carte := ia.carte(e, v)
	partir := ia.partir(e, v)
	if partir && (o.PeutSortir || o.Passage) {
		if o.Passage {
			return Choix{Type: CPassage}
		}
		return Choix{Type: CSortir}
	}
	if len(o.Sauts) > 0 {
		best, bd := o.Sauts[0], 1<<30
		for _, k := range o.Sauts {
			if carte[k] < bd {
				best, bd = k, carte[k]
			}
		}
		if !partir || !o.PeutSortir {
			return Choix{Type: CSauter, Case: best}
		}
	}
	best := Choix{}
	bs := carte[v.Case]*10 + risque(e, v, v.Case, 0)
	for k, t := range o.Trajets {
		if carte[k] == math.MaxInt32 {
			continue
		}
		s := carte[k]*10 + risque(e, v, k, len(t.Chemin))
		if t.Saut {
			s += 15
		}
		if s < bs {
			best, bs = Choix{Type: CAller, Case: k}, s
		}
	}
	return best
}

func (ia IA) combattre(e *Etat, v *Voleur, o *Options) Choix {
	for i, a := range o.Adversaires {
		if a.Type != AdvGarde || a.Dist > 1 {
			continue
		}
		g := e.Gardes[a.Indice]
		if g.Mode != Poursuite || g.Proie != v.Num {
			continue
		}
		for _, arme := range []Arme{ACouteau, AMatraque, AMainNue} {
			if contientArme(a.Armes, arme) {
				return Choix{Type: CAttaquer, Indice: i, Arme: arme}
			}
		}
	}
	return Choix{}
}
