package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"fricfrac/jeu"
)

// L'écran d'accueil : l'illustration (assets/cover.jpg) en entier, et les
// réglages dans un bandeau nuit en dessous.

func quitter() { os.Exit(0) }

func (g *Game) hauteurCouv() float64 {
	b := imgCouv.Bounds()
	return float64(LargeurEcran) * float64(b.Dy()) / float64(b.Dx())
}

func (g *Game) dessinerAccueil(dst *ebiten.Image) {
	b := imgCouv.Bounds()
	s := float64(LargeurEcran) / float64(b.Dx())
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(s, s)
	dst.DrawImage(imgCouv, op)
	hc := g.hauteurCouv()
	rect(dst, 0, hc, LargeurEcran, HauteurEcran-hc, coulNuit)
	rect(dst, 0, hc, LargeurEcran, 3, coulOr)
	rect(dst, 0, hc+6, LargeurEcran, 1, coulOrSombre)
	y := hc + 16
	ecrire(dst, "Adversaires", 40, y, face(18, policeSansGras), coulOr)
	ecrire(dst, "Votre cambrioleur", 400, y, face(18, policeSansGras), coulOr)
	ecrire(dst, jeu.NomsVoleurs[g.portrait], 400, y+52, face(17, policeItalique), coulCreme)
	ecrireDroite(dst, "D'après François Nédélec — Jeux & Stratégie", 1560, y, face(15, policeItalique), coulCreme2)
	if g.nAdv == 0 {
		ecrire(dst, "seul contre les gardes", 40, y+70, face(14, policeItalique), coulCreme2)
	}
	for i := range 6 {
		x := 570 + float64(i)*62
		jeton(dst, imgVoleurs[i], x+27, y+42, 54, ebiten.ColorScale{})
		c := coulOrSombre
		ep := 1.5
		if i == g.portrait {
			c, ep = coulOr2, 4
		}
		cadreCarre(dst, x+27, y+42, 54, ep, c)
	}
}

func (g *Game) boutonsAccueil() {
	hc := g.hauteurCouv()
	y := hc + 40
	for n := 0; n <= 5; n++ {
		n := n
		b := g.bouton(40+float64(n)*52, y, 46, 44, fmt.Sprint(n), true, func() { g.nAdv = n })
		b.Choisi = g.nAdv == n
		b.Style = BoutonDiscret
	}
	b := g.bouton(1000, y, 150, 46, "Règle", true, func() { g.regles = true })
	b.Style, b.Touche = BoutonDiscret, "R"
	b = g.bouton(1170, y-4, 260, 54, "Entrer en scène", true, func() {
		g.commencer(0)
		g.p.lancer()
	})
	b.Style, b.Touche = BoutonPrincipal, "Entrée"
	b = g.bouton(1450, y, 110, 46, "Quitter", true, quitter)
	b.Style = BoutonDiscret
}

// cliquerPortrait : le choix du pion sur l'accueil.
func (g *Game) cliquerPortrait(mx, my int) {
	y := g.hauteurCouv() + 16
	for i := range 6 {
		x := 570 + float64(i)*62
		if float64(mx) >= x && float64(mx) < x+54 && float64(my) >= y+15 && float64(my) < y+69 {
			g.portrait = i
		}
	}
}
