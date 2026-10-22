package jeu

// Les trois tables de la règle.

// Tableau 1 : table d'effraction. Portée du bruit (en cases) selon le
// nombre de tours passés sur place (1 à 8).
var tableEffraction = map[Forcable][8]int{
	FGrille:     {6, 5, 4, 3, 2, 1, 0, 0},
	FPorteExt:   {4, 3, 2, 1, 0, 0, 0, 0},
	FFenetre:    {1, 0, 0, 0, 0, 0, 0, 0},
	FPorteInt:   {2, 1, 0, 0, 0, 0, 0, 0},
	FSecretaire: {3, 2, 1, 0, 0, 0, 0, 0},
	FArmoire:    {5, 4, 3, 2, 1, 0, 0, 0},
	FCoffre:     {7, 6, 5, 4, 3, 2, 1, 0},
}

// Ce qui se force : les obstacles et les trois meubles à butin.
type Forcable uint8

const (
	FGrille Forcable = iota
	FPorteExt
	FFenetre
	FPorteInt
	FSecretaire
	FArmoire
	FCoffre
)

var NomsForcables = [...]string{"la grille", "la porte d'entrée", "la fenêtre", "la porte", "le secrétaire", "l'armoire", "le coffre-fort"}

// Bruit : portée du bruit d'une effraction de n tours (1 à 8) ; le
// passe-partout (1 ou 2 tours) compte comme autant de tours de plus, sauf
// sur une fenêtre, qui n'a pas de serrure.
func Bruit(f Forcable, tours, passe int) int {
	if f != FFenetre {
		tours += passe
	}
	return tableEffraction[f][min(max(tours, 1), 8)-1]
}

// ToursSilence : le nombre de tours pour forcer sans bruit.
func ToursSilence(f Forcable, passe int) int {
	for n := 1; n <= 8; n++ {
		if Bruit(f, n, passe) == 0 {
			return n
		}
	}
	return 8
}

// Tableau 2 : table de combat.
type Arme uint8

const (
	ASaut Arme = iota
	AMainNue
	AMatraque
	ACouteau
	ARevolver
)

var NomsArmes = [...]string{"saut", "à mains nues", "à la matraque", "au couteau", "au revolver"}

type Issue uint8

const (
	OK  Issue = iota // le défenseur est indemne
	KO               // assommé ou blessé
	Tue              // mort
)

var NomsIssues = [...]string{"indemne", "assommé", "tué"}

var tableCombat = [5][6]Issue{
	ASaut:     {OK, OK, OK, KO, KO, KO},
	AMainNue:  {OK, OK, OK, KO, KO, KO},
	AMatraque: {OK, OK, KO, KO, KO, Tue},
	ACouteau:  {OK, KO, KO, KO, Tue, Tue},
	ARevolver: {KO, KO, KO, Tue, Tue, Tue},
}

// Combat : issue d'un dé, après les retraits (distance, pare-balles).
func ResultatCombat(a Arme, de int) Issue {
	return tableCombat[a][min(max(de, 1), 6)-1]
}

// Tableau 3 : table de butin, tirée avec deux dés (dizaines, unités).
type TypeButin uint8

const (
	BRien       TypeButin = iota
	BAlarme2              // signal d'alarme, bruit à 2 dés
	BGaz                  // gaz étourdissant : 2 tours sur place
	BCharnieres           // bruit à 3 cases
	BGardes               // les gardes avancent de 3 cases
	BTombe                // bruit à 4 cases
	BChalumeau            // bruit à 6 cases
	BPassage              // passage secret vers l'extérieur
	BAlarme3              // signal d'alarme, bruit à 3 dés
	BObjet                // un objet de valeur
)

type Objet struct {
	Nom     string
	Valeur  int
	Especes bool // argent liquide et or : compté intégralement
}

type Butin struct {
	Type  TypeButin
	Objet Objet
	Texte string
}

var (
	portefeuille = Objet{"un portefeuille", 1000, true}
	piecesOr     = Objet{"des pièces d'or", 3000, true}
	bague        = Objet{"une bague", 100000, false}
	lingot       = Objet{"un lingot d'or", 5000, true}
	lettre       = Objet{"une lettre compromettante", 10000, false}
	collier      = Objet{"un collier de perles", 300000, false}
	actions      = Objet{"un portefeuille d'actions", 30000, false}
	dossier      = Objet{"un dossier secret", 50000, false}
	riviere      = Objet{"la rivière de diamants", 500000, false}
)

// TableButin : la case de la table pour un tirage de 11 à 66.
func TableButin(n int) Butin {
	obj := func(o Objet) Butin { return Butin{Type: BObjet, Objet: o, Texte: o.Nom} }
	switch n {
	case 11:
		return Butin{Type: BAlarme2, Texte: "signal d'alarme"}
	case 13:
		return Butin{Type: BGaz, Texte: "gaz étourdissant : 2 tours sur place"}
	case 15:
		return Butin{Type: BCharnieres, Texte: "les charnières grincent : bruit à 3 cases"}
	case 22:
		return Butin{Type: BGardes, Texte: "les gardes avancent de 3 cases"}
	case 24:
		return Butin{Type: BTombe, Texte: "le meuble tombe : bruit à 4 cases"}
	case 26:
		return Butin{Type: BChalumeau, Texte: "le chalumeau explose : bruit à 6 cases"}
	case 32:
		return Butin{Type: BPassage, Texte: "un passage secret vers l'extérieur"}
	case 33:
		return Butin{Type: BAlarme3, Texte: "signal d'alarme"}
	case 34, 35, 36:
		return obj(portefeuille)
	case 41, 42, 43:
		return obj(piecesOr)
	case 44:
		return obj(bague)
	case 45, 46, 51:
		return obj(lingot)
	case 52, 53, 54:
		return obj(lettre)
	case 55:
		return obj(collier)
	case 56, 61, 62:
		return obj(actions)
	case 63, 64, 65:
		return obj(dossier)
	case 66:
		return obj(riviere)
	}
	return Butin{Type: BRien, Texte: "rien"}
}

// Équipement : trois points à répartir.
type Equipement struct {
	Passe      int // passe-partout : 0, 1 tour (2 pts) ou 2 tours (3 pts)
	Chaussures bool
	PareBalles int // 0, 1 (1 pt) ou 2 (2 pts)
	Matraque   bool
	Couteau    bool
	Revolver   bool
}

const PointsEquipement = 3

func (q Equipement) Cout() int {
	n := []int{0, 2, 3}[q.Passe] + q.PareBalles
	if q.Chaussures {
		n++
	}
	if q.Matraque {
		n++
	}
	if q.Couteau {
		n += 2
	}
	if q.Revolver {
		n += 3
	}
	return n
}

func (q Equipement) Valide() bool {
	return q.Passe >= 0 && q.Passe <= 2 && q.PareBalles >= 0 && q.PareBalles <= 2 && q.Cout() <= PointsEquipement
}

// Description de l'équipement, pour le journal.
func (q Equipement) Texte() string {
	var l []string
	if q.Passe > 0 {
		l = append(l, []string{"", "passe-partout (1 tour)", "passe-partout (2 tours)"}[q.Passe])
	}
	if q.Chaussures {
		l = append(l, "chaussures silencieuses")
	}
	if q.PareBalles > 0 {
		l = append(l, []string{"", "pare-balles (1)", "pare-balles (2)"}[q.PareBalles])
	}
	if q.Matraque {
		l = append(l, "matraque")
	}
	if q.Couteau {
		l = append(l, "couteau")
	}
	if q.Revolver {
		l = append(l, "revolver")
	}
	if len(l) == 0 {
		return "rien"
	}
	s := l[0]
	for i, x := range l[1:] {
		if i == len(l)-2 {
			s += " et " + x
		} else {
			s += ", " + x
		}
	}
	return s
}
