package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type StyleBouton int

const (
	BoutonNormal    StyleBouton = iota // or sur nuit
	BoutonPrincipal                    // l'action attendue : bordeaux et or
	BoutonDiscret                      // navigation, fermeture
)

type Bouton struct {
	X, Y, W, H float64
	Texte      string
	Sous       string // seconde ligne, plus petite
	Actif      bool
	Action     func()
	Style      StyleBouton
	Taille     float64
	Choisi     bool
	Touche     string // raccourci clavier, affiché en petit
}

func (b *Bouton) Contient(x, y int) bool {
	fx, fy := float64(x), float64(y)
	return fx >= b.X && fy >= b.Y && fx < b.X+b.W && fy < b.Y+b.H
}

// tailleMinBouton : le texte des boutons n'est jamais réduit en dessous.
const tailleMinBouton = 16

func dessinerBouton(dst *ebiten.Image, b *Bouton, survol bool) {
	var fond, bord, encre color.RGBA
	switch b.Style {
	case BoutonPrincipal:
		fond, bord, encre = coulBordeaux, coulOr, coulOr2
		if survol {
			fond = melange(coulBordeaux, coulRouge, 0.35)
		}
	case BoutonDiscret:
		fond, bord, encre = coulNuit2, coulOrSombre, coulCreme
		if survol {
			fond = coulNuit3
		}
	default:
		fond, bord, encre = coulOr, coulOrSombre, coulNuit
		if survol {
			fond = coulOr2
		}
	}
	if b.Choisi {
		fond, bord, encre = coulCreme, coulBordeaux, coulBordeaux
	}
	if !b.Actif {
		fond, bord, encre = melange(fond, coulNuit, 0.6), melange(bord, coulNuit, 0.5), melange(encre, coulNuit, 0.5)
	}
	if b.Actif {
		arrondi(dst, b.X+3, b.Y+3, b.W, b.H, 4, alpha(coulNoir, 0.55), nil, 0)
	}
	arrondi(dst, b.X, b.Y, b.W, b.H, 4, fond, bord, 2)
	// filet intérieur Art déco
	if b.Style != BoutonDiscret && b.H >= 34 {
		cadre(dst, b.X+4, b.Y+4, b.W-8, b.H-8, 1, alpha(bord, 0.8))
	}
	t := b.Taille
	if t == 0 {
		t = 21
	}
	police := policeSansGras
	// la taille voulue, réduite seulement si le texte déborde du bouton
	hauteur := b.H - 6
	if b.Sous != "" {
		hauteur = b.H * 0.55
	}
	for t > tailleMinBouton && (largeur(b.Texte, face(t, police)) > b.W-20 || t*1.15 > hauteur) {
		t--
	}
	f := face(t, police)
	m := f.Metrics()
	hl := m.HAscent + m.HDescent
	if b.Touche != "" {
		ft := face(13, policeSans)
		ecrireDroite(dst, b.Touche, b.X+b.W-7, b.Y+3, ft, alpha(encre, 0.7))
	}
	if b.Sous == "" {
		ecrireCentre(dst, b.Texte, b.X+b.W/2, b.Y+(b.H-hl)/2+1, f, encre)
		return
	}
	ts := 17.0
	for ts > 14 && largeur(b.Sous, face(ts, policeSans)) > b.W-16 {
		ts--
	}
	ecrireCentre(dst, b.Texte, b.X+b.W/2, b.Y+b.H*0.3-hl/2, f, encre)
	ecrireCentre(dst, b.Sous, b.X+b.W/2, b.Y+b.H-ts*1.3-4, face(ts, policeSans), encre)
}
