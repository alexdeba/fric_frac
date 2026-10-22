package main

import (
	"sync"
	"time"

	"fricfrac/jeu"
)

// Partie : le moteur tourne dans sa propre goroutine et garde le verrou
// tant qu'il calcule ; il le rend pendant qu'il attend le joueur humain ou
// qu'une animation se joue. L'interface ne lit l'état que sous ce verrou.

type TypeReq int

const (
	ReqAucune TypeReq = iota
	ReqEquiper
	ReqAgir
)

// Demande : ce que le moteur attend du joueur humain.
type Demande struct {
	Type    TypeReq
	Num     int // change à chaque demande
	Options *jeu.Options
}

// Anim : un pion qui glisse de case en case.
type Anim struct {
	Garde  bool
	Num    int
	Chemin []int
	Debut  time.Time
	Duree  time.Duration
}

// Onde : un bruit qui s'étend.
type Onde struct {
	Case, Rayon int
	Debut       time.Time
}

// Jet : des dés à afficher.
type Jet struct {
	Joueur int
	Titre  string
	Des    []int
	Texte  string
}

type Ligne struct {
	Joueur int
	Texte  string
	Tour   int
	Fort   bool // alarme, arrestation, butin : en évidence
}

type Partie struct {
	mu sync.Mutex
	E  *jeu.Etat

	Req     Demande
	nReq    int
	reponse chan any

	Journal []Ligne
	Jets    []Jet
	Anim    *Anim
	Ondes   []Onde
	Alarme  time.Time // dernière alarme (bandeau rouge)
	Rapide  bool
	Termine bool

	// captures : l'ordinateur joue pour l'humain jusqu'à cette demande
	auto   bool
	arret  jeu.Phase
	arretN int
	ia     jeu.IA
}

func nouvellePartie(nAdv, portrait int, graine uint64) *Partie {
	noms := []string{jeu.NomsVoleurs[portrait]}
	portraits := []int{portrait}
	humains := []bool{true}
	for i := 0; len(noms) < nAdv+1; i++ {
		if i != portrait {
			noms = append(noms, jeu.NomsVoleurs[i])
			portraits = append(portraits, i)
			humains = append(humains, false)
		}
	}
	p := &Partie{reponse: make(chan any, 1)}
	p.E = jeu.NouvelEtat(noms, portraits, humains, graine)
	p.E.Ctrl = append(p.E.Ctrl, &Humain{p})
	for range nAdv {
		p.E.Ctrl = append(p.E.Ctrl, jeu.IA{})
	}
	p.E.Obs = p
	return p
}

func (p *Partie) lancer() {
	go func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.E.Jouer()
		p.Termine = true
	}()
}

// humainDehors : le joueur humain n'est plus en jeu (on accélère).
func (p *Partie) humainDehors() bool { return !p.E.Voleurs[0].Actif() }

// attendre : rend le verrou pendant d (animation, lecture).
func (p *Partie) attendre(d time.Duration) {
	if p.Rapide || p.auto || d <= 0 {
		return
	}
	if p.humainDehors() {
		d /= 3
	}
	p.mu.Unlock()
	time.Sleep(d)
	p.mu.Lock()
}

// demander : pose une demande et attend la réponse de l'interface.
func (p *Partie) demander(d Demande) any {
	p.nReq++
	d.Num = p.nReq
	p.Req = d
	p.mu.Unlock()
	r := <-p.reponse
	p.mu.Lock()
	p.Req = Demande{Num: p.nReq}
	return r
}

// repondre : appelé par l'interface, verrou tenu.
func (p *Partie) repondre(r any) {
	if p.Req.Type == ReqAucune {
		return
	}
	p.Req.Type = ReqAucune
	select {
	case p.reponse <- r:
	default:
	}
}

// pilote : l'ordinateur joue-t-il encore pour l'humain (captures) ?
func (p *Partie) pilote(ph jeu.Phase) bool {
	if !p.auto {
		return false
	}
	if ph == p.arret {
		p.arretN--
		if p.arretN <= 0 {
			p.auto = false
			return false
		}
	}
	return true
}

func (p *Partie) noter(j int, txt string, fort bool) {
	p.Journal = append(p.Journal, Ligne{Joueur: j, Texte: txt, Tour: p.E.Tour, Fort: fort})
	if len(p.Journal) > 400 {
		p.Journal = p.Journal[len(p.Journal)-400:]
	}
}

// Evenement : l'observateur du moteur.
func (p *Partie) Evenement(e *jeu.Etat, ev jeu.Ev) {
	switch ev.Type {
	case jeu.EvTexte:
		p.noter(ev.Joueur, ev.Texte, false)
	case jeu.EvTour:
		if !e.Voleurs[ev.Joueur].Humain {
			p.attendre(120 * time.Millisecond)
		}
	case jeu.EvDes:
		p.Jets = append(p.Jets, Jet{ev.Joueur, ev.Titre, ev.Des, ev.Texte})
		if len(p.Jets) > 4 {
			p.Jets = p.Jets[1:]
		}
		p.attendre(450 * time.Millisecond)
	case jeu.EvButin, jeu.EvCombat:
		p.noter(ev.Joueur, ev.Texte, true)
	case jeu.EvMouvement:
		if len(ev.Chemin) < 2 {
			return
		}
		d := time.Duration(min(110*(len(ev.Chemin)-1), 700)) * time.Millisecond
		p.Anim = &Anim{Num: ev.Joueur, Chemin: ev.Chemin, Debut: time.Now(), Duree: d}
		p.attendre(d)
		p.Anim = nil
	case jeu.EvGarde:
		d := 110 * time.Millisecond
		p.Anim = &Anim{Garde: true, Num: ev.Joueur, Chemin: ev.Chemin, Debut: time.Now(), Duree: d}
		p.attendre(d)
		p.Anim = nil
	case jeu.EvBruit:
		p.Ondes = append(p.Ondes, Onde{ev.Case, ev.Rayon, time.Now()})
		p.attendre(350 * time.Millisecond)
	case jeu.EvAlarme:
		p.Alarme = time.Now()
		p.noter(-1, ev.Texte, true)
		p.attendre(500 * time.Millisecond)
	case jeu.EvRevele:
		p.attendre(300 * time.Millisecond)
	case jeu.EvArrestation, jeu.EvSortie:
		p.attendre(500 * time.Millisecond)
	}
}

// ---------------------------------------------------------------------------
// Humain : le contrôleur du joueur humain, qui passe par l'interface.

type Humain struct{ p *Partie }

func (h *Humain) Equiper(e *jeu.Etat, v *jeu.Voleur) jeu.Equipement {
	if h.p.auto {
		return jeu.Equipement{Passe: 1, Chaussures: true}
	}
	return h.p.demander(Demande{Type: ReqEquiper}).(jeu.Equipement)
}

// distanceCombat : on ne propose le combat au joueur que si un adversaire
// est à 3 cases ou moins (sinon, dans le jardin, la question reviendrait à
// chaque tour pour un rival lointain).
const distanceCombat = 3

func (h *Humain) Agir(e *jeu.Etat, v *jeu.Voleur, o *jeu.Options) jeu.Choix {
	if o.Phase == jeu.PhaseCombat {
		proche := false
		for _, a := range o.Adversaires {
			if a.Dist <= distanceCombat {
				proche = true
			}
		}
		if !proche {
			return jeu.Choix{}
		}
	}
	if h.p.pilote(o.Phase) {
		return h.p.ia.Agir(e, v, o)
	}
	return h.p.demander(Demande{Type: ReqAgir, Options: o}).(jeu.Choix)
}
