package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"fricfrac/jeu"
)

// dessinerInfoCase : une étiquette en bas de la vue décrit la case
// survolée (le lieu, le meuble, le pion qui s'y trouve, le coût du trajet).
func (g *Game) dessinerInfoCase(dst *ebiten.Image) {
	k := g.survol
	if k < 0 || !jeu.P.Praticable(k) {
		return
	}
	e := g.p.E
	t := jeu.Majuscule(jeu.NomLieu(k))
	if m := e.MeubleSur(k); m != nil {
		switch m.Etat {
		case jeu.Cache:
			t += " · un meuble encore caché"
		case jeu.Decouvert:
			t += " · " + jeu.NomsMeubles[m.Type] + ", découvert par " + e.Voleurs[m.Decouvreur].Nom
		default:
			t += " · " + jeu.NomsMeubles[m.Type] + " (vide)"
		}
	}
	if gd := e.GardeSur(k); gd != nil {
		t += " · " + gd.Nom
		switch gd.Mode {
		case jeu.Alerte:
			t += " (alerté)"
		case jeu.Poursuite:
			t += " (en chasse)"
		}
	}
	if v := e.VoleurSur(k); v != nil {
		t += " · " + v.Nom
	}
	if o := g.options(); o != nil && o.Phase == jeu.PhaseMarche {
		if tr, ok := o.Trajets[k]; ok {
			t += fmt.Sprintf(" · %d point%s", tr.Cout, pl(tr.Cout))
			if tr.Saut {
				t += ", saut par la fenêtre"
			}
		}
	}
	f := face(18, policeSans)
	w := largeur(t, f) + 32
	x, y := 16.0, float64(HauteurEcran)-48
	arrondi(dst, x+3, y+3, w, 34, 4, alpha(coulNoir, 0.5), nil, 0)
	arrondi(dst, x, y, w, 34, 4, alpha(coulNuit, 0.92), coulOrSombre, 1.5)
	ecrire(dst, t, x+16, y+6, f, coulCreme)
}
