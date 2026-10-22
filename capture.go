package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"fricfrac/jeu"
)

// Options de test : capture d'écran et mise en place directe d'une partie.
var (
	optCapture  = flag.String("shot", "", "enregistre l'écran dans ce fichier PNG puis quitte")
	optImages   = flag.Int("frames", 30, "nombre d'images avant la capture")
	optEcran    = flag.String("ecran", "", "jeu : saute l'accueil")
	optAdv      = flag.Int("adv", 3, "nombre d'adversaires (0 à 5)")
	optPortrait = flag.Int("portrait", 0, "votre pion (0 à 5)")
	optGraine   = flag.Uint64("graine", 0, "graine du hasard (0 = aléatoire)")
	optRegle    = flag.Int("regle", 0, "afficher cette page de la règle")
	optArret    = flag.String("arret", "", "captures : l'ordinateur joue pour vous jusqu'à cette demande (equiper, entree, effraction, marche, combat, fin ; marche3 : la troisième)")
	optSurvol   = flag.String("survol", "", "captures : souris sur cette case (c,r)")
	optZoom     = flag.Float64("zoom", 0, "captures : zoom du plateau")
	optCentre   = flag.String("centre", "", "captures : vue centrée sur cette case (c,r), « moi » : votre pion")
	optProfil   = flag.String("cpuprofile", "", "profil CPU dans ce fichier")
	optPerf     = flag.Bool("perf", false, "affiche le temps moyen d'une image")
	optGrille   = flag.Bool("grille", false, "contrôle : dessine toutes les cases relevées")
)

var nomsArrets = map[string]jeu.Phase{"entree": jeu.PhaseEntree, "effraction": jeu.PhaseEffraction, "marche": jeu.PhaseMarche, "combat": jeu.PhaseCombat}

func lireCase(s string) int {
	var c, r int
	if _, err := fmt.Sscanf(s, "%d,%d", &c, &r); err != nil {
		return -1
	}
	return jeu.Indice(c, r)
}

func (g *Game) preparer() {
	if *optProfil != "" {
		f, err := os.Create(*optProfil)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}
	if *optRegle > 0 {
		g.regles, g.pageRgl = true, min(*optRegle, len(pagesRegle))-1
	}
	g.nAdv = min(max(*optAdv, 0), 5)
	g.portrait = min(max(*optPortrait, 0), 5)
	if *optEcran != "jeu" {
		return
	}
	g.commencer(*optGraine)
	if a := *optArret; a != "" && a != "equiper" {
		nom, n := strings.TrimRight(a, "0123456789"), 1
		fmt.Sscan(strings.TrimPrefix(a, nom), &n)
		g.p.auto = true
		ph, ok := nomsArrets[nom]
		if !ok {
			ph = 99 // fin : jamais demandé
		}
		g.p.arret, g.p.arretN = ph, n
	}
	if *optZoom > 0 {
		g.cam.Z = *optZoom
		g.cam.borner()
	}
	g.p.lancer()
}

// survolCapture : option -survol, la souris posée sur une case.
func (g *Game) survolCapture() {
	if k := lireCase(*optSurvol); k >= 0 {
		g.survol = k
	}
}

// capturer : après -frames images et une fois la demande attendue posée
// (ou la partie finie), enregistre l'écran.
func (g *Game) capturer(dst *ebiten.Image) {
	if *optPerf {
		g.mesurer()
	}
	if *optCapture == "" {
		return
	}
	g.images++
	if g.p != nil && g.p.auto && !g.p.Termine {
		if g.images > 40000 {
			log.Fatal("capture : demande jamais atteinte")
		}
		g.imagesPretes = 0
		return
	}
	if g.p != nil && g.p.Req.Type == ReqAucune && !g.p.Termine && *optArret != "" {
		g.imagesPretes = 0
		return
	}
	g.imagesPretes++
	if g.imagesPretes == 2 && *optCentre != "" {
		k := lireCase(*optCentre)
		if *optCentre == "moi" && g.p != nil {
			k = g.p.E.Voleurs[0].Case
		}
		if k >= 0 {
			c := jeu.Centre(k)
			g.cam.centrer(c.X, c.Y)
		}
	}
	if g.imagesPretes < *optImages {
		return
	}
	b := dst.Bounds()
	img := image.NewRGBA(b)
	dst.ReadPixels(img.Pix)
	f, err := os.Create(*optCapture)
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()
	pprof.StopCPUProfile()
	os.Exit(0)
}

// mesurer : temps moyen entre deux images, toutes les 120 images (la
// première série, qui charge les textures, n'est pas affichée).
func (g *Game) mesurer() {
	g.nPerf++
	if g.nPerf%120 == 0 {
		d := time.Since(g.tPerf)
		if g.nPerf > 120 {
			fmt.Printf("%.1f ms par image (%.0f i/s)\n", d.Seconds()*1000/120, 120/d.Seconds())
		}
		g.tPerf = time.Now()
	}
}
