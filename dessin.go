package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Dessins du jeu : hexagones de surlignage, pions, dés, ornements Art déco.

// hexagone : une case pleine (coordonnées écran, r : rayon centre-pointe).
func hexagone(dst *ebiten.Image, x, y, r float64, c color.RGBA) {
	vs := []ebiten.Vertex{sommet(x, y, c)}
	var is []uint16
	for i := 0; i <= 6; i++ {
		a := math.Pi/2 + float64(i)*math.Pi/3
		vs = append(vs, sommet(x+r*math.Cos(a), y-r*math.Sin(a), c))
		if i > 0 {
			is = append(is, 0, uint16(i), uint16(i+1))
		}
	}
	dst.DrawTriangles(vs, is, blanc, &ebiten.DrawTrianglesOptions{})
}

// contourHex : le bord d'une case.
func contourHex(dst *ebiten.Image, x, y, r, ep float64, c color.RGBA) {
	var pts [][2]float64
	for i := 0; i <= 6; i++ {
		a := math.Pi/2 + float64(i)*math.Pi/3
		pts = append(pts, [2]float64{x + r*math.Cos(a), y - r*math.Sin(a)})
	}
	polyligne(dst, pts, ep, c)
}

// dessinerDos : le dos des pions-mobilier, acajou marqueté d'un losange
// doré (on ne sait pas quel meuble se cache dessous).
func dessinerDos() *ebiten.Image {
	const t = 160
	img := ebiten.NewImage(t, t)
	p := cheminArrondi(2, 2, t-4, t-4, 14)
	remplir(img, p, coulAcajou)
	tracer(img, p, 4, coulOrSombre)
	// filets
	q := cheminArrondi(14, 14, t-28, t-28, 8)
	tracer(img, q, 2, coulOr)
	var l vector.Path
	l.MoveTo(t/2, 26)
	l.LineTo(t-26, t/2)
	l.LineTo(t/2, t-26)
	l.LineTo(26, t/2)
	l.Close()
	remplir(img, &l, coulAcajou2)
	tracer(img, &l, 3, coulOr)
	f := face(64, policeTitre)
	ecrireCentre(img, "?", t/2, t/2-40, f, coulOr2)
	return img
}

// jeton : une image de pion carrée, centrée en (x, y), de côté s.
func jeton(dst *ebiten.Image, img *ebiten.Image, x, y, s float64, cs ebiten.ColorScale) {
	b := img.Bounds()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-float64(b.Dx())/2, -float64(b.Dy())/2)
	op.GeoM.Scale(s/float64(b.Dy()), s/float64(b.Dy()))
	op.GeoM.Translate(x, y)
	op.ColorScale = cs
	dst.DrawImage(img, op)
}

// jetonTourne : un grand pion (deux cases) tourné de a radians.
func jetonTourne(dst *ebiten.Image, img *ebiten.Image, x, y, h, a float64, cs ebiten.ColorScale) {
	b := img.Bounds()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-float64(b.Dx())/2, -float64(b.Dy())/2)
	op.GeoM.Scale(h/float64(b.Dy()), h/float64(b.Dy()))
	op.GeoM.Rotate(a)
	op.GeoM.Translate(x, y)
	op.ColorScale = cs
	dst.DrawImage(img, op)
}

// de : un dé ivoire à points d'ébène.
func de(dst *ebiten.Image, x, y, s float64, v int, fond, points color.Color) {
	arrondi(dst, x+2, y+3, s, s, s*0.18, alpha(coulNoir, 0.5), nil, 0)
	arrondi(dst, x, y, s, s, s*0.18, fond, coulOrSombre, 1.6)
	r := s * 0.09
	pos := map[int][][2]float64{
		1: {{.5, .5}},
		2: {{.27, .27}, {.73, .73}},
		3: {{.25, .25}, {.5, .5}, {.75, .75}},
		4: {{.27, .27}, {.73, .27}, {.27, .73}, {.73, .73}},
		5: {{.25, .25}, {.75, .25}, {.5, .5}, {.25, .75}, {.75, .75}},
		6: {{.27, .23}, {.73, .23}, {.27, .5}, {.73, .5}, {.27, .77}, {.73, .77}},
	}[v]
	for _, q := range pos {
		disque(dst, x+q[0]*s, y+q[1]*s, r, points)
	}
}

// eventail : l'ornement Art déco des titres, un éventail de rayons dorés.
func eventail(dst *ebiten.Image, cx, cy, r float64, n int, c color.RGBA) {
	for i := range n {
		a := math.Pi + math.Pi*float64(i+1)/float64(n+1)
		trait(dst, cx+math.Cos(a)*r*0.25, cy+math.Sin(a)*r*0.25, cx+math.Cos(a)*r, cy+math.Sin(a)*r, 1.6, c)
	}
	cercle(dst, cx, cy, r*0.22, 1.6, c)
}

// filetDeco : une ligne dorée à gradins (frise Art déco) de x0 à x1.
func filetDeco(dst *ebiten.Image, x0, x1, y float64, c color.RGBA) {
	rect(dst, x0, y, x1-x0, 1.5, c)
	rect(dst, x0+18, y+4, x1-x0-36, 1, c)
	m := (x0 + x1) / 2
	for _, d := range []float64{0, 8, 16} {
		rect(dst, m-24+d, y-3-d/4, 48-2*d, 1, c)
	}
	losangeDeco(dst, x0, y+1, 5, c)
	losangeDeco(dst, x1, y+1, 5, c)
}

func losangeDeco(dst *ebiten.Image, cx, cy, r float64, c color.RGBA) {
	vs := []ebiten.Vertex{sommet(cx-r, cy, c), sommet(cx, cy-r, c), sommet(cx+r, cy, c), sommet(cx, cy+r, c)}
	dst.DrawTriangles(vs, []uint16{0, 1, 2, 0, 2, 3}, blanc, &ebiten.DrawTrianglesOptions{})
}

// cadreDeco : cadre doré double à coins coupés.
func cadreDeco(dst *ebiten.Image, x, y, w, h float64, c color.RGBA) {
	cadre(dst, x, y, w, h, 2, c)
	cadre(dst, x+5, y+5, w-10, h-10, 1, c)
	for _, p := range [][2]float64{{x, y}, {x + w, y}, {x, y + h}, {x + w, y + h}} {
		losangeDeco(dst, p[0], p[1], 6, c)
	}
}

// bulleBruit : une onde de bruit qui s'étend (t de 0 à 1).
func bulleBruit(dst *ebiten.Image, x, y, r, t float64, c color.RGBA) {
	for k := range 3 {
		u := t - float64(k)*0.18
		if u <= 0 || u > 1 {
			continue
		}
		anneau(dst, x, y, r*u, 4*(1-u)+1, alpha(c, (1-u)*0.9))
	}
}
