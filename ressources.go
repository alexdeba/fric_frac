package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"fricfrac/jeu"
)

//go:embed assets/gen/plateau.jpg assets/gen/couverture.jpg assets/gen/pions assets/polices
var fichiers embed.FS

var (
	imgPlateau *ebiten.Image // le plateau redressé
	imgCouv    *ebiten.Image // l'illustration de l'écran d'accueil
	blanc      *ebiten.Image // 3 × 3 blanc, pour les triangles

	imgVoleurs [6]*ebiten.Image
	imgGardes  [2]*ebiten.Image
	imgMeubles [6]*ebiten.Image // dans l'ordre de jeu.TypeMeuble
	imgDos     *ebiten.Image    // dos des pions-mobilier

	policeTitre, policeTexte, policeGras, policeItalique, policeSans, policeSansGras, policeMachine *Police
)

// Palette « Neuilly, 1925 » : la nuit bleu ardoise des fenêtres, l'or des
// lampes et des ferronneries Art déco, l'acajou des boiseries, le bordeaux
// des tapis et des portes, le papier crème des cartes de visite.
var (
	coulNuit     = color.RGBA{0x12, 0x17, 0x24, 0xff} // fond du panneau
	coulNuit2    = color.RGBA{0x1d, 0x26, 0x3a, 0xff}
	coulNuit3    = color.RGBA{0x2c, 0x38, 0x52, 0xff}
	coulOr       = color.RGBA{0xd9, 0xb2, 0x5f, 0xff}
	coulOr2      = color.RGBA{0xf3, 0xd9, 0x8f, 0xff} // or clair
	coulOrSombre = color.RGBA{0x8f, 0x6c, 0x2b, 0xff}
	coulCreme    = color.RGBA{0xf4, 0xe9, 0xcf, 0xff}
	coulCreme2   = color.RGBA{0xe6, 0xd5, 0xae, 0xff}
	coulAcajou   = color.RGBA{0x4a, 0x24, 0x14, 0xff} // encre des cartes
	coulAcajou2  = color.RGBA{0x7a, 0x3e, 0x22, 0xff}
	coulBordeaux = color.RGBA{0x8c, 0x22, 0x30, 0xff}
	coulRouge    = color.RGBA{0xd6, 0x2f, 0x2f, 0xff}
	coulVert     = color.RGBA{0x3f, 0x9a, 0x5c, 0xff}
	coulBleu     = color.RGBA{0x3d, 0x8f, 0xd6, 0xff} // gardes
	coulBlanc    = color.RGBA{0xff, 0xff, 0xff, 0xff}
	coulNoir     = color.RGBA{0x00, 0x00, 0x00, 0xff}
)

// couleurs des voleurs (le premier est le joueur humain)
var coulVoleurs = []color.RGBA{
	{0xe8, 0xb9, 0x3c, 0xff}, // or
	{0xd6, 0x45, 0x8f, 0xff}, // rose
	{0x4f, 0xc0, 0x8a, 0xff}, // émeraude
	{0xe0, 0x6a, 0x2c, 0xff}, // orange
	{0x9b, 0x7b, 0xe0, 0xff}, // lilas
	{0x5f, 0xc3, 0xd9, 0xff}, // turquoise
}

func chargerImage(nom string) *ebiten.Image {
	b, err := fichiers.ReadFile(nom)
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(nom, err)
	}
	return ebiten.NewImageFromImage(img)
}

// Police : une fonte, et le poids voulu pour les fontes variables.
type Police struct {
	src   *text.GoTextFaceSource
	poids float32
}

func chargerPolice(nom string, poids float32) *Police {
	b, err := fichiers.ReadFile("assets/polices/" + nom)
	if err != nil {
		log.Fatal(err)
	}
	s, err := text.NewGoTextFaceSource(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	return &Police{s, poids}
}

func chargerRessources() {
	imgPlateau = chargerImage("assets/gen/plateau.jpg")
	imgCouv = chargerImage("assets/gen/couverture.jpg")
	blanc = ebiten.NewImage(3, 3)
	blanc.Fill(color.White)
	for i := range imgVoleurs {
		imgVoleurs[i] = chargerImage(fmt.Sprintf("assets/gen/pions/voleur%d.png", i+1))
	}
	for i := range imgGardes {
		imgGardes[i] = chargerImage(fmt.Sprintf("assets/gen/pions/garde%d.png", i+1))
	}
	for i, n := range []string{"coffre", "secretaire", "gueridon", "fauteuil", "armoire", "lit"} {
		imgMeubles[i] = chargerImage("assets/gen/pions/" + n + ".png")
	}
	policeTitre = chargerPolice("Limelight-Regular.ttf", 0)
	policeTexte = chargerPolice("PlayfairDisplay.ttf", 400)
	policeGras = chargerPolice("PlayfairDisplay.ttf", 700)
	policeItalique = chargerPolice("PlayfairDisplay-Italique.ttf", 400)
	policeSans = chargerPolice("JosefinSans.ttf", 400)
	policeSansGras = chargerPolice("JosefinSans.ttf", 700)
	policeMachine = chargerPolice("SpecialElite-Regular.ttf", 0)
	imgDos = dessinerDos()
}

// ---------------------------------------------------------------- texte

type cleFace struct {
	p *Police
	t float64
}

var faces = map[cleFace]*text.GoTextFace{}

func face(taille float64, p *Police) *text.GoTextFace {
	k := cleFace{p, taille}
	if f, ok := faces[k]; ok {
		return f
	}
	f := &text.GoTextFace{Source: p.src, Size: taille}
	if p.poids > 0 {
		f.SetVariation(text.MustParseTag("wght"), p.poids)
	}
	faces[k] = f
	return f
}

func ecrire(dst *ebiten.Image, s string, x, y float64, f text.Face, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, f, op)
}

func ecrireCentre(dst *ebiten.Image, s string, cx, y float64, f text.Face, c color.Color) {
	ecrire(dst, s, cx-text.Advance(s, f)/2, y, f, c)
}

func ecrireDroite(dst *ebiten.Image, s string, x, y float64, f text.Face, c color.Color) {
	ecrire(dst, s, x-text.Advance(s, f), y, f, c)
}

// ecrireHalo : texte cerné.
func ecrireHalo(dst *ebiten.Image, s string, x, y float64, f text.Face, c, halo color.Color, ep float64) {
	for a := 0.0; a < 2*math.Pi; a += math.Pi / 6 {
		ecrire(dst, s, x+math.Cos(a)*ep, y+math.Sin(a)*ep, f, halo)
	}
	ecrire(dst, s, x, y, f, c)
}

func largeur(s string, f text.Face) float64 { return text.Advance(s, f) }

// mots : la ponctuation haute reste avec le mot qui précède.
func mots(s string) []string {
	var l []string
	for _, m := range strings.Fields(s) {
		if n := len(l); n > 0 && (m == ";" || m == ":" || m == "!" || m == "?" || m == "»" || m == "$" || m == "%" || l[n-1] == "«") {
			l[n-1] += " " + m
			continue
		}
		l = append(l, m)
	}
	return l
}

// couper : un texte en lignes de largeur maximale w.
func couper(s string, f text.Face, w float64) []string {
	var lignes []string
	for _, para := range strings.Split(s, "\n") {
		cur := ""
		for _, mot := range mots(para) {
			essai := mot
			if cur != "" {
				essai = cur + " " + mot
			}
			if text.Advance(essai, f) > w && cur != "" {
				lignes = append(lignes, cur)
				cur = mot
			} else {
				cur = essai
			}
		}
		lignes = append(lignes, cur)
	}
	return lignes
}

// paragraphe : écrit un texte coupé, renvoie la hauteur utilisée.
func paragraphe(dst *ebiten.Image, s string, x, y, w float64, f text.Face, c color.Color, inter float64) float64 {
	l := couper(s, f, w)
	for i, t := range l {
		ecrire(dst, t, x, y+float64(i)*inter, f, c)
	}
	return float64(len(l)) * inter
}

// ---------------------------------------------------------------- formes

func rgba(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func rect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	k := rgba(c)
	vs := []ebiten.Vertex{sommet(x, y, k), sommet(x+w, y, k), sommet(x, y+h, k), sommet(x+w, y+h, k)}
	dst.DrawTriangles(vs, []uint16{0, 1, 2, 1, 3, 2}, blanc, &ebiten.DrawTrianglesOptions{})
}

func cadre(dst *ebiten.Image, x, y, w, h, ep float64, c color.Color) {
	k := rgba(c)
	rect(dst, x-ep/2, y-ep/2, w+ep, ep, k)
	rect(dst, x-ep/2, y+h-ep/2, w+ep, ep, k)
	rect(dst, x-ep/2, y+ep/2, ep, h-ep, k)
	rect(dst, x+w-ep/2, y+ep/2, ep, h-ep, k)
}

// disque : plein, bord adouci d'un pixel.
func disque(dst *ebiten.Image, x, y, r float64, c color.Color) {
	k := rgba(c)
	pastille(dst, x, y, max(r-0.5, 0.5), k)
	anneau(dst, x, y, r-0.25, 0.5, k)
}

func cercle(dst *ebiten.Image, x, y, r, ep float64, c color.Color) {
	anneau(dst, x, y, r, ep, rgba(c))
}

func trait(dst *ebiten.Image, x0, y0, x1, y1, ep float64, c color.Color) {
	polyligne(dst, [][2]float64{{x0, y0}, {x1, y1}}, ep, rgba(c))
}

// arrondi : rectangle aux coins arrondis.
func cheminArrondi(x, y, w, h, r float64) *vector.Path {
	var p vector.Path
	r = min(r, w/2, h/2)
	p.MoveTo(float32(x+r), float32(y))
	p.LineTo(float32(x+w-r), float32(y))
	p.ArcTo(float32(x+w), float32(y), float32(x+w), float32(y+r), float32(r))
	p.LineTo(float32(x+w), float32(y+h-r))
	p.ArcTo(float32(x+w), float32(y+h), float32(x+w-r), float32(y+h), float32(r))
	p.LineTo(float32(x+r), float32(y+h))
	p.ArcTo(float32(x), float32(y+h), float32(x), float32(y+h-r), float32(r))
	p.LineTo(float32(x), float32(y+r))
	p.ArcTo(float32(x), float32(y), float32(x+r), float32(y), float32(r))
	p.Close()
	return &p
}

func remplir(dst *ebiten.Image, p *vector.Path, c color.Color) {
	op := &vector.FillOptions{}
	dop := &vector.DrawPathOptions{AntiAlias: true}
	dop.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, p, op, dop)
}

func tracer(dst *ebiten.Image, p *vector.Path, ep float64, c color.Color) {
	op := &vector.StrokeOptions{Width: float32(ep), LineJoin: vector.LineJoinRound, LineCap: vector.LineCapRound}
	dop := &vector.DrawPathOptions{AntiAlias: true}
	dop.ColorScale.ScaleWithColor(c)
	vector.StrokePath(dst, p, op, dop)
}

// arrondi : rectangle aux coins arrondis, rendu une fois (anticrénelé)
// dans une image gardée en cache, puis simplement recopié.
func arrondi(dst *ebiten.Image, x, y, w, h, r float64, fond color.Color, bord color.Color, ep float64) {
	k := cleArrondi{w: int(math.Round(w)), h: int(math.Round(h)), r: int(math.Round(r)), ep: int(math.Round(ep * 2))}
	if fond != nil {
		k.fond, k.aFond = rgba(fond), true
	}
	if bord != nil && ep > 0 {
		k.bord, k.aBord = rgba(bord), true
	}
	img, ok := arrondis[k]
	if !ok {
		if len(arrondis) > 2000 { // tailles très variées : on repart à zéro
			for _, i := range arrondis {
				i.Deallocate()
			}
			arrondis = map[cleArrondi]*ebiten.Image{}
		}
		m := int(math.Ceil(ep)) + 2
		img = ebiten.NewImage(k.w+2*m, k.h+2*m)
		p := cheminArrondi(float64(m), float64(m), float64(k.w), float64(k.h), r)
		if fond != nil {
			remplir(img, p, fond)
		}
		if bord != nil && ep > 0 {
			tracer(img, p, ep, bord)
		}
		arrondis[k] = img
	}
	m := float64(int(math.Ceil(ep)) + 2)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(math.Round(x)-m, math.Round(y)-m)
	dst.DrawImage(img, op)
}

type cleArrondi struct {
	w, h, r, ep  int
	fond, bord   color.RGBA
	aFond, aBord bool
}

var arrondis = map[cleArrondi]*ebiten.Image{}

func alpha(c color.RGBA, a float64) color.RGBA {
	return color.RGBA{uint8(float64(c.R) * a), uint8(float64(c.G) * a), uint8(float64(c.B) * a), uint8(float64(c.A) * a)}
}

func melange(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), m(a.A, b.A)}
}

func dist(a, b jeu.Point) float64 { return math.Hypot(a.X-b.X, a.Y-b.Y) }

// ---------------------------------------------------------------- triangles

// Formes dessinées directement en triangles, sans le rasteriseur
// vectoriel : pour tout ce qui se redessine à chaque image sur la carte
// (anneaux qui pulsent, trajets), bien plus rapide.

func sommet(x, y float64, c color.RGBA) ebiten.Vertex {
	return ebiten.Vertex{DstX: float32(x), DstY: float32(y), SrcX: 1, SrcY: 1,
		ColorR: float32(c.R) / 255, ColorG: float32(c.G) / 255, ColorB: float32(c.B) / 255, ColorA: float32(c.A) / 255}
}

// anneau : cercle d'épaisseur ep, bords adoucis par un fondu d'un pixel.
func anneau(dst *ebiten.Image, x, y, r, ep float64, c color.RGBA) {
	n := max(int(r*0.7), 24)
	var vs []ebiten.Vertex
	var is []uint16
	transp := color.RGBA{}
	rayons := []float64{r - ep/2 - 1, r - ep/2, r + ep/2, r + ep/2 + 1}
	couls := []color.RGBA{transp, c, c, transp}
	for i := 0; i <= n; i++ {
		a := 2 * math.Pi * float64(i) / float64(n)
		ca, sa := math.Cos(a), math.Sin(a)
		for k, rr := range rayons {
			vs = append(vs, sommet(x+rr*ca, y+rr*sa, couls[k]))
		}
		if i > 0 {
			b := uint16(4 * (i - 1))
			for k := uint16(0); k < 3; k++ {
				is = append(is, b+k, b+k+1, b+4+k, b+k+1, b+5+k, b+4+k)
			}
		}
	}
	dst.DrawTriangles(vs, is, blanc, &ebiten.DrawTrianglesOptions{})
}

// polyligne : trait épais le long de points (écran), joints arrondis.
func polyligne(dst *ebiten.Image, pts [][2]float64, ep float64, c color.RGBA) {
	var vs []ebiten.Vertex
	var is []uint16
	for i := 1; i < len(pts); i++ {
		x0, y0, x1, y1 := pts[i-1][0], pts[i-1][1], pts[i][0], pts[i][1]
		l := math.Hypot(x1-x0, y1-y0)
		if l == 0 {
			continue
		}
		nx, ny := -(y1-y0)/l*ep/2, (x1-x0)/l*ep/2
		b := uint16(len(vs))
		vs = append(vs, sommet(x0+nx, y0+ny, c), sommet(x0-nx, y0-ny, c), sommet(x1+nx, y1+ny, c), sommet(x1-nx, y1-ny, c))
		is = append(is, b, b+1, b+2, b+1, b+3, b+2)
		if len(vs) > 60000 {
			break
		}
	}
	dst.DrawTriangles(vs, is, blanc, &ebiten.DrawTrianglesOptions{})
	for _, p := range pts[1 : len(pts)-1] {
		pastille(dst, p[0], p[1], ep/2, c)
	}
}

// pastille : disque plein en triangles.
func pastille(dst *ebiten.Image, x, y, r float64, c color.RGBA) {
	n := max(int(r), 10)
	vs := []ebiten.Vertex{sommet(x, y, c)}
	var is []uint16
	for i := 0; i <= n; i++ {
		a := 2 * math.Pi * float64(i) / float64(n)
		vs = append(vs, sommet(x+r*math.Cos(a), y+r*math.Sin(a), c))
		if i > 0 {
			is = append(is, 0, uint16(i), uint16(i+1))
		}
	}
	dst.DrawTriangles(vs, is, blanc, &ebiten.DrawTrianglesOptions{})
}
