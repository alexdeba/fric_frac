// Fric-Frac — d'après le jeu de François Nédélec paru en encart dans
// Jeux & Stratégie (assets/Regles.jpg, assets/Encart.jpg). Un cambrioleur
// humain et jusqu'à cinq rivaux menés par l'ordinateur se disputent le
// butin d'un hôtel particulier de Neuilly gardé par deux gardes.
package main

import (
	"flag"
	"log"
	"math/rand/v2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fricfrac/jeu"
)

const (
	LargeurEcran = 1600
	HauteurEcran = 1000
	LargVue      = 1100.0 // le plateau
	PanneauX     = LargVue
	LargPanneau  = LargeurEcran - PanneauX
)

type Ecran int

const (
	EcranAccueil Ecran = iota
	EcranJeu
)

type Game struct {
	ecran    Ecran
	nAdv     int
	portrait int
	p        *Partie
	numReq   int

	cam        Camera
	glisse     bool
	glisseMini bool
	gx, gy     float64
	gcx, gcy   float64
	survol     int // case sous la souris, -1

	equip   jeu.Equipement // écran d'équipement
	regles  bool
	pageRgl int
	defil   int // défilement du journal

	boutons  []*Bouton
	libelles []libelle // légendes des lignes d'action du panneau
	yJournal float64
	hCarte   float64 // hauteur de la carte de consigne
	mx, my   int
	t0       time.Time
	images   int
	// captures et mesures
	imagesPretes int
	nPerf        int
	tPerf        time.Time
	suiviX       int // dernière case suivie par la caméra
}

func nouveauJeu() *Game {
	g := &Game{nAdv: 3, t0: time.Now(), survol: -1, suiviX: -1}
	g.cam.toutVoir()
	return g
}

func (g *Game) Layout(int, int) (int, int) { return LargeurEcran, HauteurEcran }

func (g *Game) temps() float64 { return time.Since(g.t0).Seconds() }

func (g *Game) commencer(graine uint64) {
	if graine == 0 {
		graine = rand.Uint64()
	}
	g.p = nouvellePartie(g.nAdv, g.portrait, graine)
	g.ecran = EcranJeu
	g.numReq = -1
	g.cam.toutVoir()
}

func (g *Game) Update() error {
	g.mx, g.my = ebiten.CursorPosition()
	clic := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if g.p != nil {
		g.p.mu.Lock()
		defer g.p.mu.Unlock()
		if g.p.Req.Num != g.numReq {
			g.numReq = g.p.Req.Num
			g.nouvelleDemande()
		}
	}
	g.construireBoutons()
	if clic {
		for _, b := range g.boutons {
			if b.Actif && b.Contient(g.mx, g.my) {
				b.Action()
				return nil
			}
		}
	}
	if g.regles {
		g.touchesRegle()
		return nil
	}
	if g.ecran == EcranAccueil {
		if clic {
			g.cliquerPortrait(g.mx, g.my)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.commencer(0)
			g.p.lancer()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.regles = true
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.regles = true
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.p.Rapide = !g.p.Rapide
	}
	g.touchesDemande()
	surMini := g.majMiniCarte()
	if !surMini {
		g.cameraClavierSouris()
	}
	g.suivreAction()
	g.survol = -1
	if !surMini && g.mx < int(LargVue) && g.p.Req.Type != ReqEquiper && !g.finAffichee() {
		x, y := g.cam.carte(float64(g.mx), float64(g.my))
		g.survol = jeu.CaseSous(x, y)
	}
	g.survolCapture()
	if clic && !surMini && g.survol >= 0 {
		g.cliquerCase(g.survol)
	}
	return nil
}

// nouvelleDemande : remise à zéro de la saisie à chaque demande du moteur.
func (g *Game) nouvelleDemande() {
	if g.p.Req.Type == ReqEquiper {
		g.equip = jeu.Equipement{}
	}
	if g.p.Req.Type != ReqAucune {
		g.defil = 0
	}
}

func (g *Game) finAffichee() bool { return g.p != nil && g.p.Termine }

func (g *Game) cameraClavierSouris() {
	sx, sy := float64(g.mx), float64(g.my)
	if _, dy := ebiten.Wheel(); dy != 0 {
		if sx < LargVue {
			g.cam.zoomer(1+0.12*dy, sx, sy)
		} else {
			g.defil = max(g.defil+int(dy), 0)
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		if sx < LargVue {
			g.glisse, g.gx, g.gy, g.gcx, g.gcy = true, sx, sy, g.cam.X, g.cam.Y
		}
	}
	if g.glisse {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
			g.cam.X = g.gcx - (sx-g.gx)/g.cam.Z
			g.cam.Y = g.gcy - (sy-g.gy)/g.cam.Z
			g.cam.borner()
		} else {
			g.glisse = false
		}
	}
	v := 14 / g.cam.Z
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.cam.X -= v
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.cam.X += v
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.cam.Y -= v
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.cam.Y += v
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) || inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		g.cam.zoomer(1.25, LargVue/2, HauteurEcran/2)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) || inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		g.cam.zoomer(1/1.25, LargVue/2, HauteurEcran/2)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDigit0) || inpututil.IsKeyJustPressed(ebiten.KeyNumpad0) || inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		g.cam.toutVoir()
	}
	g.cam.borner()
}

// suivreAction : quand le plateau est zoomé, la vue ne suit un pion que
// s'il vient de bouger hors du champ ; elle ne reprend jamais la main sur
// la vue que le joueur a choisie (glisser, mini-carte).
func (g *Game) suivreAction() {
	a := g.p.Anim
	if a == nil || !g.zoomee() || g.glisse || g.glisseMini {
		return
	}
	k := a.Chemin[len(a.Chemin)-1]
	if k == g.suiviX {
		return
	}
	g.suiviX = k
	c := jeu.Centre(k)
	sx, sy := g.cam.ecran(c.X, c.Y)
	if sx < 60 || sy < 60 || sx > LargVue-60 || sy > HauteurEcran-60 {
		g.cam.centrer(c.X, c.Y)
	}
}

func (g *Game) Draw(dst *ebiten.Image) {
	if g.p != nil {
		g.p.mu.Lock()
		defer g.p.mu.Unlock()
	}
	dst.Fill(coulNuit)
	switch g.ecran {
	case EcranAccueil:
		g.dessinerAccueil(dst)
	case EcranJeu:
		g.dessinerPlateau(dst)
		g.dessinerMiniCarte(dst)
		if g.p.Req.Type == ReqEquiper {
			g.dessinerEquipement(dst)
		}
		if g.finAffichee() {
			g.dessinerFin(dst)
		}
		g.dessinerPanneau(dst)
	}
	if g.regles {
		g.dessinerRegles(dst)
	}
	for _, b := range g.boutons {
		dessinerBouton(dst, b, b.Actif && b.Contient(g.mx, g.my))
	}
	g.capturer(dst)
}

func (g *Game) bouton(x, y, w, h float64, t string, actif bool, f func()) *Bouton {
	b := &Bouton{X: x, Y: y, W: w, H: h, Texte: t, Actif: actif, Action: f}
	g.boutons = append(g.boutons, b)
	return b
}

func (g *Game) construireBoutons() {
	g.boutons = g.boutons[:0]
	g.libelles = g.libelles[:0]
	if g.regles {
		g.boutonsRegles()
		return
	}
	switch g.ecran {
	case EcranAccueil:
		g.boutonsAccueil()
	case EcranJeu:
		g.boutonsJeu()
	}
}

var optPleinEcran = flag.Bool("fullscreen", false, "démarre en plein écran (F11 bascule)")

func main() {
	flag.Parse()
	chargerRessources()
	g := nouveauJeu()
	g.preparer()
	ebiten.SetWindowSize(LargeurEcran, HauteurEcran)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Fric-Frac — casse à Neuilly")
	ebiten.SetFullscreen(*optPleinEcran)
	if *optPerf {
		ebiten.SetVsyncEnabled(false) // mesurer le coût réel d'une image
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
