package jeu

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

type TypeMeuble uint8

const (
	MCoffre TypeMeuble = iota
	MSecretaire
	MGueridon
	MFauteuil
	MArmoire
	MLit
)

var NomsMeubles = [...]string{"un coffre-fort", "un secrétaire", "un guéridon", "un fauteuil", "une armoire", "un lit"}

// Grand : armoires et lits tiennent deux cases.
func (t TypeMeuble) Grand() bool { return t == MArmoire || t == MLit }

// Forcable : la ligne de la table d'effraction, et le nombre de tirages de
// butin (1 secrétaire, 2 armoire, 3 coffre-fort) ; 0 : rien à prendre.
func (t TypeMeuble) Forcable() (Forcable, int) {
	switch t {
	case MSecretaire:
		return FSecretaire, 1
	case MArmoire:
		return FArmoire, 2
	case MCoffre:
		return FCoffre, 3
	}
	return 0, 0
}

type EtatMeuble uint8

const (
	Cache     EtatMeuble = iota // face cachée
	Decouvert                   // retourné ; seul son découvreur peut le forcer
	Force                       // forcé : vidé
	Vide                        // le découvreur s'est éloigné sans le forcer
)

type EtatDormeur uint8

const (
	Endormi EtatDormeur = iota
	Eveille
	Assomme
	Mort
)

type Meuble struct {
	Num        int
	Type       TypeMeuble
	Cases      []int // une ou deux cases
	Dir        int   // grands meubles : direction de Cases[0] vers Cases[1]
	Etat       EtatMeuble
	Decouvreur int // numéro du voleur, -1
	// lit : le dormeur, tête sur Cases[1] (côté oreiller)
	Dormeur   EtatDormeur
	DormeurKO int // tour où il reprend ses esprits
}

func (m *Meuble) Tete() int { return m.Cases[len(m.Cases)-1] }

// Adjacent : la case k touche-t-elle le meuble ?
func (m *Meuble) Adjacent(k int) bool {
	if k < 0 {
		return false
	}
	for _, c := range m.Cases {
		if Dist(c, k) == 1 {
			return true
		}
	}
	return false
}

func (m *Meuble) Occupe(k int) bool {
	for _, c := range m.Cases {
		if c == k {
			return true
		}
	}
	return false
}

type Statut uint8

const (
	Attente Statut = iota // hors du plateau, pas encore entré
	EnJeu
	Sorti
	Arrete
	Abattu
)

type Voleur struct {
	Num      int
	Nom      string
	Portrait int
	Humain   bool
	Equip    Equipement
	Statut   Statut
	Case     int // -1 : hors du plateau
	Cote     int // sur une case-seuil : la région d'où l'on vient

	Immobile int // tours à passer sur place (assommé, saut, gaz)
	Travail  *Travail
	Ouverts  map[int]bool // obstacles forcés ce tour-ci

	Butins   []Objet
	Vu       bool // vu par un garde ou un dormeur
	AAssomme bool // a assommé un garde ou un dormeur
	ATue     bool
	Passage  int  // meuble au passage secret, -1
	Repli    bool // l'ordinateur a décidé de repartir
	Score    int
	TourFin  int
}

// Travail : une effraction de plusieurs tours en cours.
type Travail struct {
	Eff      Effraction
	Tours    int
	Restants int
}

func (v *Voleur) Actif() bool { return v.Statut == Attente || v.Statut == EnJeu }

// SurPlateau : le pion est sur une case.
func (v *Voleur) SurPlateau() bool { return v.Statut == EnJeu && v.Case >= 0 }

// Valeur : le butin estimé au prix de la table.
func (v *Voleur) Valeur() int {
	n := 0
	for _, o := range v.Butins {
		n += o.Valeur
	}
	return n
}

// Diviseur : barème du receleur.
func (v *Voleur) Diviseur() int {
	switch {
	case v.ATue:
		return 10
	case v.AAssomme || v.Vu:
		return 5
	}
	return 1
}

// Recel : la somme touchée chez le receleur.
func (v *Voleur) Recel() int {
	esp, autres := 0, 0
	for _, o := range v.Butins {
		if o.Especes {
			esp += o.Valeur
		} else {
			autres += o.Valeur
		}
	}
	return esp + autres/v.Diviseur()
}

type ModeGarde uint8

const (
	Ronde     ModeGarde = iota // déplacement au hasard
	Alerte                     // vers la case suspecte
	Poursuite                  // derrière un voleur vu
)

type Garde struct {
	Num   int
	Nom   string
	Base  int
	Dirs  [6]int // direction pour chaque face du dé
	Case  int
	Dir   int
	Mode  ModeGarde
	Cible int // case suspecte (alerte, poursuite)
	Proie int // voleur poursuivi
	KO    int // tour où il reprend ses esprits
	Mort  bool
}

func (g *Garde) Actif(tour int) bool { return !g.Mort && tour >= g.KO }

type Cadavre struct {
	Case    int
	Trouve  bool
	Libelle string
}

type Etat struct {
	Voleurs []*Voleur
	Gardes  [2]*Garde
	Meubles []*Meuble
	Tour    int
	Courant int
	PM      int // points de déplacement : le nombre de joueurs

	Ronde    bool // les gardes ont commencé leur ronde
	Alarme   bool // l'alarme a été donnée : deux cases par tour
	Cadavres []*Cadavre

	Ctrl []Controleur
	Obs  Observateur
	Fini bool

	rng *rand.Rand
}

// TourMax : au lever du jour, ceux qui sont encore dans la propriété sont
// pris.
const TourMax = 150

var NomsVoleurs = []string{"Le Gentleman", "La Féline", "Le Grand Gégène", "Bébert le Foulard", "Mimi la Gouaille", "Monsieur Paul"}
var NomsGardes = [2]string{"le garde Lebœuf", "le garde Planchet"}

// NouvelEtat : portraits[i] est le pion du joueur i (0 : l'humain s'il y
// en a un).
func NouvelEtat(noms []string, portraits []int, humains []bool, graine uint64) *Etat {
	e := &Etat{rng: rand.New(rand.NewPCG(graine, graine^0x5fa1c3)), PM: len(noms)}
	for i, n := range noms {
		e.Voleurs = append(e.Voleurs, &Voleur{Num: i, Nom: n, Portrait: portraits[i], Humain: humains[i], Case: -1, Cote: RHors, Passage: -1})
	}
	e.Gardes[0] = &Garde{Num: 0, Nom: NomsGardes[0], Base: P.BaseA, Dirs: P.DirsA, Case: P.BaseA, Proie: -1}
	e.Gardes[1] = &Garde{Num: 1, Nom: NomsGardes[1], Base: P.BaseB, Dirs: P.DirsB, Case: P.BaseB, Proie: -1}
	e.placerMeubles()
	return e
}

func (e *Etat) De() int { return e.rng.IntN(6) + 1 }

// Rng : le hasard de la partie (pour l'IA).
func (e *Etat) Rng() *rand.Rand { return e.rng }

// MeubleSur : le meuble qui occupe la case, nil.
func (e *Etat) MeubleSur(k int) *Meuble {
	for _, m := range e.Meubles {
		if m.Occupe(k) {
			return m
		}
	}
	return nil
}

func (e *Etat) GardeSur(k int) *Garde {
	for _, g := range e.Gardes {
		if !g.Mort && g.Case == k {
			return g
		}
	}
	return nil
}

func (e *Etat) VoleurSur(k int) *Voleur {
	for _, v := range e.Voleurs {
		if v.SurPlateau() && v.Case == k {
			return v
		}
	}
	return nil
}

// ---------------------------------------------------------------- meubles

// placerMeubles : les 36 pions-mobilier, face cachée, au hasard dans les
// pièces : jamais sur une porte ni contre elle, jamais deux meubles sur des
// cases adjacentes, ni sur les bases des gardes.
func (e *Etat) placerMeubles() {
	var types []TypeMeuble
	for _, t := range []TypeMeuble{MCoffre, MSecretaire, MGueridon, MFauteuil, MArmoire, MLit} {
		for range 6 {
			types = append(types, t)
		}
	}
	// les grands d'abord : ils ont besoin de plus de place
	for essai := 0; ; essai++ {
		e.Meubles = e.Meubles[:0]
		ok := true
		e.rng.Shuffle(len(types), func(i, j int) { types[i], types[j] = types[j], types[i] })
		ordre := append([]TypeMeuble(nil), types...)
		grands := 0
		for i, t := range ordre {
			if t.Grand() {
				ordre[grands], ordre[i] = ordre[i], ordre[grands]
				grands++
			}
		}
		for _, t := range ordre {
			if !e.poserMeuble(t) {
				ok = false
				break
			}
		}
		if ok {
			break
		}
		if essai > 200 {
			panic("placement des meubles impossible")
		}
	}
	for i, m := range e.Meubles {
		m.Num = i
	}
}

// caseMeublable : case d'une pièce, pas contre une porte, sans meuble
// autour.
func (e *Etat) caseMeublable(k int) bool {
	if k < 0 || P.Terrain[k] != TPiece || k == P.BaseA || k == P.BaseB {
		return false
	}
	// pas contre une porte (une fenêtre peut être masquée par un meuble)
	for d := range 6 {
		v := Voisin(k, d)
		if o := P.ObsLien[k][d]; o >= 0 && P.Obstacles[o].Type != OFenetre {
			return false
		}
		if v >= 0 && P.Terrain[v] == TSeuil && P.Obstacles[P.ObsCase[v]].Type != OFenetre {
			return false
		}
	}
	for _, m := range e.Meubles {
		for _, c := range m.Cases {
			if c == k || Dist(c, k) == 1 {
				return false
			}
		}
	}
	return true
}

func (e *Etat) poserMeuble(t TypeMeuble) bool {
	// parmi toutes les places possibles, de préférence celles qui en
	// condamnent le moins (sinon les 36 pions ne tiennent pas)
	type place struct{ a, b, d int }
	var l []place
	for k := range NbCases {
		if !e.caseMeublable(k) {
			continue
		}
		if !t.Grand() {
			l = append(l, place{k, -1, -1})
			continue
		}
		for d := range 6 {
			k2 := Voisin(k, d)
			if e.caseMeublable(k2) && P.Piece[k2] == P.Piece[k] && P.Liens[k][d] == LLibre {
				l = append(l, place{k, k2, d})
			}
		}
	}
	if len(l) == 0 {
		return false
	}
	libres := func(k int) int {
		n := 0
		for d := range 6 {
			if v := Voisin(k, d); v >= 0 && e.caseMeublable(v) {
				n++
			}
		}
		return n
	}
	best, bs := l[0], 1<<30
	for range 12 {
		p := l[e.rng.IntN(len(l))]
		s := libres(p.a)
		if p.b >= 0 {
			s += libres(p.b)
		}
		if s < bs {
			best, bs = p, s
		}
	}
	m := &Meuble{Type: t, Cases: []int{best.a}, Dir: best.d, Decouvreur: -1}
	if best.b >= 0 {
		m.Cases = append(m.Cases, best.b)
	}
	e.Meubles = append(e.Meubles, m)
	return true
}

// ---------------------------------------------------------------- journal

func (e *Etat) evenement(ev Ev) {
	if e.Obs != nil {
		e.Obs.Evenement(e, ev)
	}
}

func (e *Etat) dire(j int, format string, args ...any) {
	e.evenement(Ev{Type: EvTexte, Joueur: j, Texte: fmt.Sprintf(format, args...)})
}

// Francs : 12 000 F
func Francs(n int) string {
	s := fmt.Sprint(n)
	out := ""
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += " "
		}
		out += string(ch)
	}
	return out + " F"
}

// Le : un nom de voleur au milieu d'une phrase (« La Féline » → « la
// Féline »).
func Le(nom string) string {
	for _, a := range []string{"Le ", "La "} {
		if strings.HasPrefix(nom, a) {
			return strings.ToLower(a[:1]) + nom[1:]
		}
	}
	return nom
}

// De : « de » devant un nom de voleur (du Gentleman, de la Féline).
func De(nom string) string {
	switch {
	case strings.HasPrefix(nom, "Le "):
		return "du " + nom[3:]
	case strings.HasPrefix(nom, "La "):
		return "de la " + nom[3:]
	}
	return "de " + nom
}
