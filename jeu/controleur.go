package jeu

// Controleur : les décisions d'un joueur. Le moteur pose des Options pour
// la phase en cours ; le contrôleur renvoie un Choix, que le moteur
// vérifie (un choix illégal vaut « passer »).
type Controleur interface {
	// Equiper : les trois points d'équipement, en début de partie.
	Equiper(e *Etat, v *Voleur) Equipement
	// Agir : une action de la phase en cours, ou Passer.
	Agir(e *Etat, v *Voleur, o *Options) Choix
}

type Phase uint8

const (
	PhaseEntree     Phase = iota // hors du plateau : choisir sa case-mur
	PhaseEffraction              // forcer obstacles et meubles (avant de bouger)
	PhaseMarche                  // se déplacer
	PhaseCombat                  // attaquer un adversaire de la pièce
)

var NomsPhases = [...]string{"entrée", "effraction", "déplacement", "combat"}

// Effraction : un obstacle ou un meuble que le voleur peut forcer d'ici.
type Effraction struct {
	Forcable Forcable
	Obstacle int // indice dans P.Obstacles, -1
	Meuble   int // indice dans Etat.Meubles, -1
	Bruits   [8]int
}

// Trajet : un déplacement possible ce tour.
type Trajet struct {
	Chemin []int // cases traversées, départ exclu
	Cout   int   // points de déplacement
	Saut   bool  // sortie par une fenêtre sans arbre : on saute
}

type TypeAdv uint8

const (
	AdvGarde TypeAdv = iota
	AdvVoleur
	AdvDormeur
)

// Adversaire : un pion qu'on peut attaquer d'ici.
type Adversaire struct {
	Type   TypeAdv
	Indice int // garde, voleur ou meuble (lit)
	Case   int
	Dist   int
	Armes  []Arme
}

type Options struct {
	Phase Phase
	PM    int // points de déplacement restants

	Entrees []int // cases-mur libres (la grille comprise)

	Effractions []Effraction

	Trajets    map[int]Trajet // par case d'arrivée
	Sauts      []int          // depuis une case-mur : cases du jardin
	PeutSortir bool           // sur une case-mur : quitter la propriété
	Passage    bool           // le passage secret est à portée

	Adversaires []Adversaire
}

type TypeChoix uint8

const (
	Passer TypeChoix = iota
	CEntrer
	CForcer
	CAller
	CSauter
	CSortir
	CPassage
	CAttaquer
)

type Choix struct {
	Type   TypeChoix
	Case   int // entrée, déplacement, saut
	Indice int // effraction, adversaire
	Tours  int // effraction : tours sur place
	Arme   Arme
}

// Observateur : suit la partie (journal, animations).
type Observateur interface {
	Evenement(e *Etat, ev Ev)
}

type TypeEv uint8

const (
	EvTexte TypeEv = iota
	EvTour
	EvDes
	EvMouvement // voleur : Chemin, départ compris
	EvGarde     // un garde avance d'une case : Joueur = garde, Chemin
	EvBruit     // Case, Rayon
	EvAlarme
	EvRevele // Indice = meuble
	EvButin  // Texte
	EvCombat
	EvSortie
	EvArrestation
	EvFin
)

type Ev struct {
	Type   TypeEv
	Joueur int
	Texte  string
	Titre  string
	Des    []int
	Chemin []int
	Case   int
	Rayon  int
	Indice int
}
