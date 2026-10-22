package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fricfrac/jeu"
)

// Le plateau : la photo redressée de l'encart, zoomable, avec les pions,
// les cases jouables surlignées et les bruits qui se propagent.

type Camera struct{ X, Y, Z float64 }

// zMin : tout le plateau, avec une petite marge (les pions posés sur le
// mur d'enceinte débordent du bord).
func (c *Camera) zMin() float64 {
	return 0.975 * min(LargVue/float64(jeu.LargPlateau), float64(HauteurEcran)/float64(jeu.HautPlateau))
}

func (c *Camera) toutVoir() {
	c.Z = c.zMin()
	c.X = -(LargVue/c.Z - jeu.LargPlateau) / 2
	c.Y = -(HauteurEcran/c.Z - jeu.HautPlateau) / 2
}

func (c *Camera) borner() {
	c.Z = min(max(c.Z, c.zMin()), 1.4)
	w, h := LargVue/c.Z, HauteurEcran/c.Z
	if w >= jeu.LargPlateau {
		c.X = -(w - jeu.LargPlateau) / 2
	} else {
		c.X = min(max(c.X, 0), jeu.LargPlateau-w)
	}
	if h >= jeu.HautPlateau {
		c.Y = -(h - jeu.HautPlateau) / 2
	} else {
		c.Y = min(max(c.Y, 0), jeu.HautPlateau-h)
	}
}

func (c *Camera) zoomer(f, sx, sy float64) {
	mx, my := c.X+sx/c.Z, c.Y+sy/c.Z
	c.Z *= f
	c.borner()
	c.X, c.Y = mx-sx/c.Z, my-sy/c.Z
	c.borner()
}

func (c *Camera) centrer(x, y float64) {
	c.X, c.Y = x-LargVue/2/c.Z, y-HauteurEcran/2/c.Z
	c.borner()
}

func (c *Camera) ecran(x, y float64) (float64, float64)   { return (x - c.X) * c.Z, (y - c.Y) * c.Z }
func (c *Camera) carte(sx, sy float64) (float64, float64) { return c.X + sx/c.Z, c.Y + sy/c.Z }

func (g *Game) ecranCase(k int) (float64, float64) {
	p := jeu.Centre(k)
	return g.cam.ecran(p.X, p.Y)
}

// ------------------------------------------------------------ saisie

// options : les options de la demande en cours (nil sinon).
func (g *Game) options() *jeu.Options {
	if g.p == nil || g.p.Req.Type != ReqAgir {
		return nil
	}
	return g.p.Req.Options
}

func (g *Game) cliquerCase(k int) {
	o := g.options()
	if o == nil {
		return
	}
	switch o.Phase {
	case jeu.PhaseEntree:
		for _, c := range o.Entrees {
			if c == k {
				g.p.repondre(jeu.Choix{Type: jeu.CEntrer, Case: k})
			}
		}
	case jeu.PhaseMarche:
		if _, ok := o.Trajets[k]; ok {
			g.p.repondre(jeu.Choix{Type: jeu.CAller, Case: k})
			return
		}
		for _, c := range o.Sauts {
			if c == k {
				g.p.repondre(jeu.Choix{Type: jeu.CSauter, Case: k})
			}
		}
	}
}

// touchesDemande : Entrée ou Échap passent la phase en cours.
func (g *Game) touchesDemande() {
	o := g.options()
	if o == nil || o.Phase == jeu.PhaseEntree {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.p.repondre(jeu.Choix{Type: jeu.Passer})
	}
}

// ------------------------------------------------------------ dessin

// pulsation : de 0 à 1 et retour, environ une fois par seconde.
func (g *Game) pulsation() float64 { return 0.5 + 0.5*math.Sin(g.temps()*5) }

func (g *Game) dessinerPlateau(dst *ebiten.Image) {
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-g.cam.X, -g.cam.Y)
	op.GeoM.Scale(g.cam.Z, g.cam.Z)
	dst.DrawImage(imgPlateau, op)
	if g.p == nil {
		return
	}
	e := g.p.E
	r := jeu.RayonCase * g.cam.Z
	if *optGrille {
		for k := range jeu.NbCases {
			x, y := g.ecranCase(k)
			contourHex(dst, x, y, r, 1.5, coulRouge)
		}
	}
	g.dessinerSurlignage(dst, r)
	g.dessinerMeubles(dst, r)
	for _, c := range e.Cadavres {
		x, y := g.ecranCase(c.Case)
		trait(dst, x-r*0.35, y-r*0.35, x+r*0.35, y+r*0.35, max(3, r*0.12), coulRouge)
		trait(dst, x-r*0.35, y+r*0.35, x+r*0.35, y-r*0.35, max(3, r*0.12), coulRouge)
	}
	g.dessinerGardes(dst, r)
	g.dessinerVoleurs(dst, r)
	g.dessinerOndes(dst)
	g.dessinerChemin(dst, r)
	// ce qu'on peut forcer : par-dessus les pions
	if o := g.options(); o != nil && o.Phase == jeu.PhaseEffraction {
		for _, eff := range o.Effractions {
			x, y := g.posEffraction(eff)
			cercle(dst, x, y, r*0.55, max(4, r*0.12), alpha(coulRouge, 0.55+0.45*g.pulsation()))
			cercle(dst, x, y, r*0.55+max(4, r*0.12), 1.5, coulOr2)
		}
	}
	// alarme : liseré rouge qui palpite
	if t := time.Since(g.p.Alarme).Seconds(); t < 2 && !g.p.Alarme.IsZero() {
		a := (1 - t/2) * (0.6 + 0.4*g.pulsation())
		for i := range 4 {
			cadre(dst, 4+float64(i)*5, 4+float64(i)*5, LargVue-8-float64(i)*10, HauteurEcran-8-float64(i)*10, 5, alpha(coulRouge, a*(1-float64(i)*0.22)))
		}
	}
	g.dessinerInfoCase(dst)
}

// dessinerSurlignage : ce que la phase en cours permet.
func (g *Game) dessinerSurlignage(dst *ebiten.Image, r float64) {
	o := g.options()
	if o == nil {
		return
	}
	pu := g.pulsation()
	or := alpha(coulOr2, 0.38+0.17*pu)
	ep := max(1.5, r*0.06)
	switch o.Phase {
	case jeu.PhaseEntree:
		for _, k := range o.Entrees {
			x, y := g.ecranCase(k)
			c := or
			if jeu.P.Terrain[k] == jeu.TGrille {
				c = alpha(coulVert, 0.45+0.2*pu)
			}
			hexagone(dst, x, y, r*0.92, c)
			contourHex(dst, x, y, r*0.92, ep, coulOr2)
			if k == g.survol {
				contourHex(dst, x, y, r*0.92, 3*ep, coulBordeaux)
			}
		}
	case jeu.PhaseMarche:
		for k, t := range o.Trajets {
			x, y := g.ecranCase(k)
			c := or
			if t.Saut {
				c = alpha(coulRouge, 0.30+0.15*pu)
			}
			hexagone(dst, x, y, r*0.88, c)
			contourHex(dst, x, y, r*0.88, ep, alpha(coulOr2, 0.9))
		}
		for _, k := range o.Sauts {
			x, y := g.ecranCase(k)
			hexagone(dst, x, y, r*0.88, alpha(coulRouge, 0.38+0.17*pu))
			contourHex(dst, x, y, r*0.88, ep, coulOr2)
		}
	case jeu.PhaseCombat:
		for _, a := range o.Adversaires {
			x, y := g.ecranCase(a.Case)
			cercle(dst, x, y, r*0.95, 4, alpha(coulRouge, 0.6+0.4*pu))
		}
	}
	// le pion du joueur en cours
	v := g.p.E.Voleurs[0]
	if v.Case >= 0 {
		x, y := g.ecranCase(v.Case)
		cercle(dst, x, y, r*(0.98+0.06*pu), 3, coulOr2)
	}
}

// posEffraction : où dessiner l'obstacle ou le meuble à forcer.
func (g *Game) posEffraction(eff jeu.Effraction) (float64, float64) {
	if eff.Meuble >= 0 {
		m := g.p.E.Meubles[eff.Meuble]
		x0, y0 := g.ecranCase(m.Cases[0])
		x1, y1 := g.ecranCase(m.Cases[len(m.Cases)-1])
		return (x0 + x1) / 2, (y0 + y1) / 2
	}
	c := jeu.P.Obstacles[eff.Obstacle].Centre
	return g.cam.ecran(c.X, c.Y)
}

func (g *Game) dessinerMeubles(dst *ebiten.Image, r float64) {
	e := g.p.E
	s := r * 1.42 // côté d'un pion
	for _, m := range e.Meubles {
		x0, y0 := g.ecranCase(m.Cases[0])
		x, y := x0, y0
		if len(m.Cases) == 2 {
			x1, y1 := g.ecranCase(m.Cases[1])
			x, y = (x0+x1)/2, (y0+y1)/2
		}
		var cs ebiten.ColorScale
		img := imgMeubles[m.Type]
		if m.Etat == jeu.Cache {
			img = imgDos
		}
		if m.Etat == jeu.Force || m.Etat == jeu.Vide {
			cs.Scale(0.45, 0.42, 0.42, 1)
		}
		// ombre portée
		if len(m.Cases) == 2 {
			a := float64(m.Dir) * math.Pi / 3
			if m.Etat == jeu.Cache {
				// le dos, étiré sur les deux cases
				jetonTourne(dst, imgDos, x+3, y+4, s, a, ombre())
				for _, k := range m.Cases {
					xx, yy := g.ecranCase(k)
					jeton(dst, imgDos, xx, yy, s, cs)
				}
			} else {
				jetonTourne(dst, img, x+3, y+4, s, a, ombre())
				jetonTourne(dst, img, x, y, s, a, cs)
			}
		} else {
			jeton(dst, img, x+3, y+4, s, ombre())
			jeton(dst, img, x, y, s, cs)
		}
		switch {
		case m.Etat == jeu.Decouvert && m.Decouvreur >= 0:
			// le découvreur : un liseré à sa couleur
			cercle(dst, x, y, s*0.2, 4, coulVoleurs[g.p.E.Voleurs[m.Decouvreur].Portrait])
		case m.Etat == jeu.Force || m.Etat == jeu.Vide:
			ecrireCentre(dst, "vide", x, y-s*0.18, face(max(11, s*0.26), policeSansGras), coulCreme)
		}
		if m.Type == jeu.MLit && m.Etat != jeu.Cache {
			tx, ty := g.ecranCase(m.Tete())
			switch m.Dormeur {
			case jeu.Eveille:
				disque(dst, tx, ty-s*0.45, s*0.2, coulRouge)
				ecrireCentre(dst, "!", tx, ty-s*0.45-s*0.17, face(s*0.3, policeSansGras), coulBlanc)
			case jeu.Assomme:
				ecrireCentre(dst, "zzz", tx, ty-s*0.6, face(max(11, s*0.25), policeSansGras), coulCreme)
			}
		}
	}
}

func ombre() ebiten.ColorScale {
	var cs ebiten.ColorScale
	cs.Scale(0, 0, 0, 0.45)
	return cs
}

// position : la case d'un pion, interpolée pendant une animation.
func (g *Game) position(garde bool, num, k int) (float64, float64) {
	a := g.p.Anim
	if a != nil && a.Garde == garde && a.Num == num && len(a.Chemin) >= 2 {
		t := min(time.Since(a.Debut).Seconds()/a.Duree.Seconds(), 1)
		f := t * float64(len(a.Chemin)-1)
		i := min(int(f), len(a.Chemin)-2)
		u := f - float64(i)
		x0, y0 := g.ecranCase(a.Chemin[i])
		x1, y1 := g.ecranCase(a.Chemin[i+1])
		return x0 + (x1-x0)*u, y0 + (y1-y0)*u
	}
	return g.ecranCase(k)
}

func (g *Game) dessinerGardes(dst *ebiten.Image, r float64) {
	e := g.p.E
	s := r * 1.3
	for _, gd := range e.Gardes {
		if gd.Mort {
			continue
		}
		x, y := g.position(true, gd.Num, gd.Case)
		var cs ebiten.ColorScale
		bord := coulBleu
		if !gd.Actif(e.Tour) {
			cs.Scale(0.5, 0.5, 0.55, 1)
			bord = melange(coulBleu, coulNuit, 0.5)
		} else if gd.Mode != jeu.Ronde {
			bord = melange(coulRouge, coulOr2, 0.5*g.pulsation())
		}
		jeton(dst, imgGardes[gd.Num], x+3, y+4, s, ombre())
		jeton(dst, imgGardes[gd.Num], x, y, s, cs)
		cadreCarre(dst, x, y, s, max(3, r*0.08), bord)
		if !gd.Actif(e.Tour) {
			ecrireCentre(dst, "KO", x, y-s*0.2, face(max(12, s*0.3), policeSansGras), coulOr2)
		}
	}
}

// cadreCarre : liseré autour d'un pion carré.
func cadreCarre(dst *ebiten.Image, x, y, s, ep float64, c interface{ RGBA() (r, g, b, a uint32) }) {
	cadre(dst, x-s/2, y-s/2, s, s, ep, c)
}

func (g *Game) dessinerVoleurs(dst *ebiten.Image, r float64) {
	e := g.p.E
	s := r * 1.22
	for i := len(e.Voleurs) - 1; i >= 0; i-- { // l'humain par-dessus
		v := e.Voleurs[i]
		if !v.SurPlateau() {
			continue
		}
		x, y := g.position(false, v.Num, v.Case)
		var cs ebiten.ColorScale
		if v.Immobile > 0 {
			cs.Scale(0.6, 0.6, 0.6, 1)
		}
		jeton(dst, imgVoleurs[v.Portrait], x+3, y+4, s, ombre())
		jeton(dst, imgVoleurs[v.Portrait], x, y, s, cs)
		ep := max(3, r*0.09)
		if e.Courant == v.Num && !g.finAffichee() {
			cadreCarre(dst, x, y, s+ep*2, ep, alpha(coulOr2, 0.5+0.5*g.pulsation()))
		}
		cadreCarre(dst, x, y, s, ep, coulVoleurs[v.Portrait])
		if v.Travail != nil {
			ecrireCentre(dst, "…", x+s*0.4, y-s*0.75, face(max(14, s*0.4), policeSansGras), coulOr2)
		}
	}
}

func (g *Game) dessinerOndes(dst *ebiten.Image) {
	var reste []Onde
	for _, o := range g.p.Ondes {
		t := time.Since(o.Debut).Seconds() / 1.3
		if t > 1 {
			continue
		}
		reste = append(reste, o)
		x, y := g.ecranCase(o.Case)
		bulleBruit(dst, x, y, (float64(o.Rayon)+0.5)*jeu.GDX*g.cam.Z, t, coulOr2)
	}
	g.p.Ondes = reste
}

// dessinerChemin : le trajet vers la case survolée, et la portée du bruit
// des pas (sans chaussures silencieuses).
func (g *Game) dessinerChemin(dst *ebiten.Image, r float64) {
	o := g.options()
	if o == nil || o.Phase != jeu.PhaseMarche || g.survol < 0 {
		return
	}
	t, ok := o.Trajets[g.survol]
	v := g.p.E.Voleurs[0]
	if !ok {
		return
	}
	pts := [][2]float64{}
	x, y := g.ecranCase(v.Case)
	pts = append(pts, [2]float64{x, y})
	for _, k := range t.Chemin {
		x, y := g.ecranCase(k)
		pts = append(pts, [2]float64{x, y})
	}
	polyligne(dst, pts, max(4, r*0.14), alpha(coulBordeaux, 0.9))
	polyligne(dst, pts, max(2, r*0.06), coulOr2)
	fx, fy := pts[len(pts)-1][0], pts[len(pts)-1][1]
	contourHex(dst, fx, fy, r*0.92, 3, coulOr2)
	if !v.Equip.Chaussures {
		n := len(t.Chemin)
		cercle(dst, fx, fy, (float64(n)+0.5)*jeu.GDX*g.cam.Z, 2, alpha(coulOr2, 0.8))
	}
}
