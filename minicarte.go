package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fricfrac/jeu"
)

// Zoom à la barre d'espace et mini-carte : quand le plateau est zoomé, un
// médaillon en haut à droite de la vue montre toute la propriété, les
// pions et le cadre de ce qu'on voit ; un clic ou un glisser dedans y
// déplace la vue.

const (
	zoomEspace = 0.85
	miniW      = 250.0
	miniMarge  = 14.0
)

var imgMini *ebiten.Image // le plateau réduit, préparé une fois

func miniH() float64 { return miniW * float64(jeu.HautPlateau) / float64(jeu.LargPlateau) }
func miniX() float64 { return LargVue - miniW - miniMarge }

const miniY = miniMarge

func (g *Game) zoomee() bool { return g.cam.Z > g.cam.zMin()*1.01 }

func (g *Game) miniVisible() bool {
	return g.p != nil && g.zoomee() && g.p.Req.Type != ReqEquiper && !g.finAffichee()
}

func (g *Game) surMini() bool {
	if !g.miniVisible() {
		return false
	}
	x, y := float64(g.mx), float64(g.my)
	return x >= miniX() && x < miniX()+miniW && y >= miniY && y < miniY+miniH()
}

// basculerZoom : vue d'ensemble ↔ zoom sur la souris (ou sur votre pion
// si la souris n'est pas sur le plateau).
func (g *Game) basculerZoom() {
	if g.zoomee() {
		g.cam.toutVoir()
		return
	}
	var x, y float64
	if sx, sy := float64(g.mx), float64(g.my); sx >= 0 && sx < LargVue && sy >= 0 && sy < HauteurEcran {
		x, y = g.cam.carte(sx, sy)
	} else if v := g.p.E.Voleurs[0]; v.Case >= 0 {
		c := jeu.Centre(v.Case)
		x, y = c.X, c.Y
	} else {
		x, y = jeu.LargPlateau/2, jeu.HautPlateau/2
	}
	g.cam.Z = zoomEspace
	g.cam.centrer(x, y)
}

func (g *Game) majMiniCarte() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.basculerZoom()
	}
	sur := g.surMini()
	if sur && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.glisseMini = true
	}
	if g.glisseMini {
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			g.glisseMini = false
		} else {
			f := float64(jeu.LargPlateau) / miniW
			g.cam.centrer((float64(g.mx)-miniX())*f, (float64(g.my)-miniY)*f)
		}
	}
	return sur || g.glisseMini
}

// preparerMini : le plateau réduit par moitiés successives (filtre
// linéaire propre), une fois pour toutes.
func preparerMini() {
	img := imgPlateau
	for img.Bounds().Dx() > int(miniW)*2 {
		b := img.Bounds()
		r := ebiten.NewImage(b.Dx()/2, b.Dy()/2)
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Scale(0.5, 0.5)
		r.DrawImage(img, op)
		img = r
	}
	imgMini = img
}

func (g *Game) dessinerMiniCarte(dst *ebiten.Image) {
	if !g.miniVisible() {
		return
	}
	if imgMini == nil {
		preparerMini()
	}
	x0, y0, w, h := miniX(), miniY, miniW, miniH()
	arrondi(dst, x0-6+4, y0-6+5, w+12, h+12, 6, alpha(coulNoir, 0.6), nil, 0)
	arrondi(dst, x0-6, y0-6, w+12, h+12, 6, coulNuit, coulOr, 2)
	s := w / float64(imgMini.Bounds().Dx())
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(x0, y0)
	dst.DrawImage(imgMini, op)
	f := w / float64(jeu.LargPlateau)
	for _, gd := range g.p.E.Gardes {
		if !gd.Mort {
			c := jeu.Centre(gd.Case)
			disque(dst, x0+c.X*f, y0+c.Y*f, 4.5, coulNuit)
			disque(dst, x0+c.X*f, y0+c.Y*f, 3.4, coulBleu)
		}
	}
	for _, v := range g.p.E.Voleurs {
		if v.SurPlateau() {
			c := jeu.Centre(v.Case)
			disque(dst, x0+c.X*f, y0+c.Y*f, 5, coulNuit)
			disque(dst, x0+c.X*f, y0+c.Y*f, 3.8, coulVoleurs[v.Portrait])
		}
	}
	vx, vy := g.cam.X*f, g.cam.Y*f
	vw, vh := LargVue/g.cam.Z*f, HauteurEcran/g.cam.Z*f
	rx, ry := max(x0+vx, x0), max(y0+vy, y0)
	rw, rh := min(x0+vx+vw, x0+w)-rx, min(y0+vy+vh, y0+h)-ry
	rect(dst, rx, ry, rw, rh, alpha(coulOr2, 0.15))
	cadre(dst, rx, ry, rw, rh, 2.5, coulOr2)
}
