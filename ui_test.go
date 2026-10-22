package main

import (
	"testing"
	"time"

	"fricfrac/jeu"
)

// Le joueur humain répond par l'interface : une partie se joue en
// répondant aux demandes comme le ferait un clic.
func TestHumainParDemandes(t *testing.T) {
	for graine := uint64(1); graine <= 5; graine++ {
		p := nouvellePartie(2, 0, graine)
		p.Rapide = true
		p.lancer()
		ia := jeu.IA{}
		limite := time.Now().Add(20 * time.Second)
		demandes := map[jeu.Phase]int{}
		for {
			p.mu.Lock()
			if p.Termine {
				p.mu.Unlock()
				break
			}
			switch p.Req.Type {
			case ReqEquiper:
				p.repondre(jeu.Equipement{Chaussures: true, Couteau: true})
			case ReqAgir:
				o := p.Req.Options
				demandes[o.Phase]++
				p.repondre(ia.Agir(p.E, p.E.Voleurs[0], o))
			}
			p.mu.Unlock()
			if time.Now().After(limite) {
				t.Fatal("partie trop longue")
			}
			time.Sleep(50 * time.Microsecond)
		}
		v := p.E.Voleurs[0]
		if v.Actif() {
			t.Fatalf("graine %d : le voleur est encore en jeu", graine)
		}
		if !v.Equip.Couteau || !v.Equip.Chaussures {
			t.Fatalf("équipement perdu : %+v", v.Equip)
		}
		if demandes[jeu.PhaseEntree] != 1 || demandes[jeu.PhaseMarche] == 0 {
			t.Fatalf("demandes : %v", demandes)
		}
		t.Logf("graine %d : statut %d, %s en %d tours ; demandes %v", graine, v.Statut, jeu.Francs(v.Recel()), v.TourFin, demandes)
	}
}

// Barre d'espace : zoom sur la souris puis retour à la vue d'ensemble.
func TestZoomEspace(t *testing.T) {
	g := nouveauJeu()
	g.commencer(1)
	g.mx, g.my = 300, 400
	x0, y0 := g.cam.carte(300, 400)
	g.basculerZoom()
	if !g.zoomee() || g.cam.Z != zoomEspace {
		t.Fatalf("pas de zoom : %+v", g.cam)
	}
	sx, sy := g.cam.ecran(x0, y0)
	if sx < 0 || sy < 0 || sx > LargVue || sy > HauteurEcran {
		t.Errorf("le point visé sort de la vue : %v %v", sx, sy)
	}
	if !g.miniVisible() {
		t.Error("mini-carte absente une fois zoomé")
	}
	g.basculerZoom()
	if g.zoomee() || g.miniVisible() {
		t.Fatalf("retour à la vue d'ensemble raté : %+v", g.cam)
	}
}

// La case sous la souris : les centres des cases tombent sur elles.
func TestCaseSous(t *testing.T) {
	for k := range jeu.NbCases {
		c := jeu.Centre(k)
		if c.X < 0 || c.Y < 0 || c.X > jeu.LargPlateau || c.Y > jeu.HautPlateau {
			continue
		}
		if jeu.CaseSous(c.X+5, c.Y-5) != k {
			t.Fatalf("case %d mal retrouvée", k)
		}
	}
}

// Aucun bouton de l'interface n'a un texte réduit sous la taille minimale.
func TestTailleBoutons(t *testing.T) {
	chargerRessources()
	g := nouveauJeu()
	verifier := func(quoi string) {
		for _, b := range g.boutons {
			if b.Texte == "" {
				continue
			}
			taille := b.Taille
			if taille == 0 {
				taille = 21
			}
			if largeur(b.Texte, face(tailleMinBouton, policeSansGras)) > b.W-20 {
				t.Errorf("%s : « %s » ne tient pas à %d px (bouton de %.0f px)", quoi, b.Texte, tailleMinBouton, b.W)
			}
		}
	}
	g.construireBoutons()
	verifier("accueil")
	g.regles = true
	g.construireBoutons()
	verifier("règle")
	g.regles = false
	g.commencer(3)
	g.p.Req = Demande{Type: ReqEquiper}
	g.construireBoutons()
	verifier("équipement")
	eff := jeu.Effraction{Forcable: jeu.FCoffre, Obstacle: -1, Meuble: 0}
	for n := range 8 {
		eff.Bruits[n] = jeu.Bruit(jeu.FCoffre, n+1, 0)
	}
	g.p.Req = Demande{Type: ReqAgir, Options: &jeu.Options{Phase: jeu.PhaseEffraction, Effractions: []jeu.Effraction{eff, eff, eff}}}
	g.construireBoutons()
	verifier("effraction")
	adv := jeu.Adversaire{Type: jeu.AdvGarde, Dist: 1, Armes: []jeu.Arme{jeu.AMainNue, jeu.AMatraque, jeu.ACouteau, jeu.ARevolver}}
	g.p.Req = Demande{Type: ReqAgir, Options: &jeu.Options{Phase: jeu.PhaseCombat, Adversaires: []jeu.Adversaire{adv, adv}}}
	g.construireBoutons()
	verifier("combat")
}
