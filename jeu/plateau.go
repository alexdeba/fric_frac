package jeu

import "math"

// Le plateau : la villa de Neuilly et son jardin, sur une grille
// d'hexagones relevée sur la photo de l'encart (plateau_donnees.go, écrit
// par outils/releve). Une case est repérée par son indice r*Cols + c.

type Terrain uint8

const (
	THors     Terrain = iota
	TEnceinte         // case-mur, à cheval sur le mur d'enceinte
	TGrille           // case-mur de la grille d'entrée
	TJardin           // pelouse, allées, terrasses, perron
	TArbre            // jardin sous un arbre
	TCoupee           // case coupée par un mur plein : inutilisable
	TSeuil            // case coupée par une porte ou une fenêtre : l'obstacle
	TPiece            // case d'une pièce de la villa
)

// Liaison entre deux cases voisines
type Lien uint8

const (
	LLibre Lien = iota
	LMur
	LPorte
	LFenetre
)

const NbCases = Cols * Rangs

// Régions : le numéro de la pièce (0…), ou l'une des valeurs suivantes.
const (
	RDehors = -1 // jardin et cases-mur
	RSeuil  = -2 // case-obstacle, entre deux régions
	RHors   = -3
)

// Directions : E, SE, SO, O, NO, NE.
var nomsDirections = [6]string{"est", "sud-est", "sud-ouest", "ouest", "nord-ouest", "nord-est"}

var dirPair = [6][2]int{{1, 0}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}, {0, -1}}
var dirImpair = [6][2]int{{1, 0}, {1, 1}, {0, 1}, {-1, 0}, {0, -1}, {1, -1}}

func Opposee(d int) int { return (d + 3) % 6 }

type TypeObstacle uint8

const (
	OGrille TypeObstacle = iota
	OPorteExt
	OFenetre
	OPorteInt
)

var NomsObstacles = [...]string{"la grille", "la porte d'entrée", "la fenêtre", "la porte"}

// Obstacle : une porte, une fenêtre ou la grille. Il occupe une case-seuil
// (Case ≥ 0 : la case coupée par la porte) ou se trouve entre deux cases
// voisines A et B (Case = -1 ; A → B dans la direction Dir).
type Obstacle struct {
	Type   TypeObstacle
	Case   int
	A, B   int
	Dir    int
	Cotes  [2]int // régions de part et d'autre
	Arbre  bool   // fenêtre près d'un arbre : on y grimpe sans peine
	Centre Point
}

type Point struct{ X, Y float64 }

type Plateau struct {
	Terrain [NbCases]Terrain
	Piece   [NbCases]int // -1 hors de la villa
	Liens   [NbCases][6]Lien

	Obstacles []Obstacle
	ObsCase   [NbCases]int    // obstacle de la case-seuil, -1
	ObsLien   [NbCases][6]int // obstacle entre la case et son voisin, -1

	BaseA, BaseB int
	// numéros autour des bases : direction prise sur chaque face du dé
	DirsA, DirsB [6]int
	Enceinte     []int // cases-mur (grille comprise)
	Grilles      []int
}

var P = chargerPlateau()

func Indice(c, r int) int { return r*Cols + c }
func ColRang(k int) (int, int) {
	return k % Cols, k / Cols
}

// Voisin : la case voisine dans la direction d, -1 hors de la grille.
func Voisin(k, d int) int {
	c, r := ColRang(k)
	v := dirPair[d]
	if r%2 != 0 {
		v = dirImpair[d]
	}
	c, r = c+v[0], r+v[1]
	if c < 0 || r < 0 || c >= Cols || r >= Rangs {
		return -1
	}
	return Indice(c, r)
}

// Direction : direction de k vers son voisin v, -1 sinon.
func Direction(k, v int) int {
	for d := range 6 {
		if Voisin(k, d) == v {
			return d
		}
	}
	return -1
}

// Centre : centre de la case en pixels de l'image du plateau.
func Centre(k int) Point {
	c, r := ColRang(k)
	x := GX0 + float64(c)*GDX
	if r%2 != 0 {
		x += GDX / 2
	}
	return Point{x, GY0 + float64(r)*GDY}
}

// RayonCase : du centre à une pointe.
const RayonCase = GDY / 1.5

// CaseSous : la case dont le centre est le plus proche du point, -1 hors.
func CaseSous(x, y float64) int {
	r := int(math.Round((y - GY0) / GDY))
	best, bd := -1, math.MaxFloat64
	for rr := r - 1; rr <= r+1; rr++ {
		if rr < 0 || rr >= Rangs {
			continue
		}
		for c := range Cols {
			k := Indice(c, rr)
			p := Centre(k)
			if d := math.Hypot(p.X-x, p.Y-y); d < bd {
				best, bd = k, d
			}
		}
	}
	if bd > RayonCase {
		return -1
	}
	return best
}

// Dist : distance en cases (sans tenir compte des murs).
func Dist(a, b int) int {
	ca, ra := ColRang(a)
	cb, rb := ColRang(b)
	// coordonnées cubiques (rangées impaires décalées à droite)
	xa := ca - (ra-(ra&1))/2
	xb := cb - (rb-(rb&1))/2
	za, zb := ra, rb
	ya, yb := -xa-za, -xb-zb
	return max(abs(xa-xb), abs(ya-yb), abs(za-zb))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Region : la pièce de la case, RDehors (jardin, mur d'enceinte), RSeuil
// ou RHors.
func (p *Plateau) Region(k int) int {
	switch p.Terrain[k] {
	case TPiece:
		return p.Piece[k]
	case TJardin, TArbre, TEnceinte, TGrille:
		return RDehors
	case TSeuil:
		return RSeuil
	}
	return RHors
}

// DansVilla : pièce ou seuil intérieur.
func (p *Plateau) DansVilla(k int) bool {
	if p.Terrain[k] == TPiece {
		return true
	}
	if p.Terrain[k] == TSeuil {
		o := p.Obstacles[p.ObsCase[k]]
		return o.Cotes[0] >= 0 || o.Cotes[1] >= 0
	}
	return false
}

// MemeRegion : deux cases sont-elles dans la même pièce (ou toutes deux
// dehors) ? Une case-seuil appartient aux deux régions qu'elle sépare.
func (p *Plateau) MemeRegion(a, b int) bool {
	ra, rb := p.regions(a), p.regions(b)
	for _, x := range ra {
		for _, y := range rb {
			if x == y && x != RHors {
				return true
			}
		}
	}
	return false
}

func (p *Plateau) regions(k int) []int {
	if p.Terrain[k] == TSeuil {
		o := p.Obstacles[p.ObsCase[k]]
		return o.Cotes[:]
	}
	return []int{p.Region(k)}
}

// Praticable : une case où un pion peut se tenir.
func (p *Plateau) Praticable(k int) bool {
	switch p.Terrain[k] {
	case THors, TCoupee:
		return false
	}
	return true
}

// lettres du terrain de plateau_donnees.go
func chargerPlateau() *Plateau {
	p := &Plateau{}
	for k := range NbCases {
		p.Piece[k] = -1
		p.ObsCase[k] = -1
		for d := range 6 {
			p.ObsLien[k][d] = -1
		}
	}
	for r := range Rangs {
		for c := range Cols {
			k := Indice(c, r)
			ch := terrainBrut[r][c]
			switch {
			case ch >= 'A' && ch <= 'Z':
				p.Terrain[k] = TPiece
				p.Piece[k] = int(ch - 'A')
			case ch == '#':
				p.Terrain[k] = TEnceinte
			case ch == '=':
				p.Terrain[k] = TGrille
			case ch == ',':
				p.Terrain[k] = TJardin
			case ch == '*':
				p.Terrain[k] = TArbre
			case ch == '%':
				p.Terrain[k] = TCoupee
			case ch == '|' || ch == '~':
				p.Terrain[k] = TSeuil
			default:
				p.Terrain[k] = THors
			}
			if p.Terrain[k] == TEnceinte || p.Terrain[k] == TGrille {
				p.Enceinte = append(p.Enceinte, k)
			}
			if p.Terrain[k] == TGrille {
				p.Grilles = append(p.Grilles, k)
			}
			for d := range 3 {
				l := map[byte]Lien{'.': LLibre, '#': LMur, '|': LPorte, '~': LFenetre}[liensBrut[r][3*c+d]]
				p.Liens[k][d] = l
				if v := Voisin(k, d); v >= 0 {
					p.Liens[v][Opposee(d)] = l
				}
			}
		}
	}
	for k := range NbCases {
		for d := range 6 {
			if Voisin(k, d) < 0 {
				p.Liens[k][d] = LMur
			}
		}
	}
	p.relierObstacles()
	p.BaseA, p.BaseB = Indice(11, 13), Indice(11, 19)
	// autour de A : 1 nord-ouest, 2 nord-est, 3 est, 4 sud-est, 5 sud-ouest,
	// 6 ouest ; autour de B : 1 sud-est, 2 sud-ouest, 3 ouest, 4 nord-ouest,
	// 5 nord-est, 6 est (relevé sur le plateau)
	p.DirsA = [6]int{4, 5, 0, 1, 2, 3}
	p.DirsB = [6]int{1, 2, 3, 4, 5, 0}
	return p
}

// relierObstacles : une case-seuil par obstacle, une liaison porte ou
// fenêtre par obstacle, et la grille.
func (p *Plateau) relierObstacles() {
	ajouter := func(o Obstacle) int {
		p.Obstacles = append(p.Obstacles, o)
		return len(p.Obstacles) - 1
	}
	for k := range NbCases {
		c, r := ColRang(k)
		switch p.Terrain[k] {
		case TSeuil:
			o := Obstacle{Case: k, A: -1, B: -1, Dir: -1, Cotes: [2]int{RHors, RHors}, Centre: Centre(k)}
			for d := range 6 {
				v := Voisin(k, d)
				if v < 0 || p.Liens[k][d] != LLibre {
					continue
				}
				reg := p.Region(v)
				if reg == RSeuil || reg == RHors {
					continue
				}
				if o.Cotes[0] == RHors {
					o.Cotes[0] = reg
				} else if o.Cotes[0] != reg {
					o.Cotes[1] = reg
				}
			}
			fen := terrainBrut[r][c] == '~'
			o.Type = typeObstacle(fen, o.Cotes)
			p.ObsCase[k] = ajouter(o)
		case TGrille:
			o := Obstacle{Type: OGrille, Case: k, A: -1, B: -1, Dir: -1, Cotes: [2]int{RHors, RDehors}, Centre: Centre(k)}
			p.ObsCase[k] = ajouter(o)
		}
	}
	for k := range NbCases {
		for d := range 3 {
			v := Voisin(k, d)
			l := p.Liens[k][d]
			if v < 0 || (l != LPorte && l != LFenetre) || !p.Praticable(k) || !p.Praticable(v) {
				continue
			}
			ra, rb := p.Region(k), p.Region(v)
			if ra == RSeuil || rb == RSeuil {
				// la case-seuil voisine porte déjà l'obstacle
				p.Liens[k][d], p.Liens[v][Opposee(d)] = LMur, LMur
				continue
			}
			pa, pb := Centre(k), Centre(v)
			o := Obstacle{Case: -1, A: k, B: v, Dir: d, Cotes: [2]int{ra, rb}, Centre: Point{(pa.X + pb.X) / 2, (pa.Y + pb.Y) / 2}}
			o.Type = typeObstacle(l == LFenetre, o.Cotes)
			i := ajouter(o)
			p.ObsLien[k][d], p.ObsLien[v][Opposee(d)] = i, i
		}
	}
	// fenêtres près d'un arbre
	for i := range p.Obstacles {
		o := &p.Obstacles[i]
		if o.Type != OFenetre {
			continue
		}
		cases := []int{o.Case}
		if o.Case < 0 {
			cases = []int{o.A, o.B}
		}
		for _, k := range cases {
			for d := range 6 {
				if v := Voisin(k, d); v >= 0 && p.Terrain[v] == TArbre {
					o.Arbre = true
				}
			}
		}
	}
}

func typeObstacle(fenetre bool, cotes [2]int) TypeObstacle {
	switch {
	case fenetre:
		return OFenetre
	case cotes[0] == RDehors || cotes[1] == RDehors:
		return OPorteExt
	}
	return OPorteInt
}

// Exterieur : l'obstacle donne-t-il sur le jardin ?
func (o *Obstacle) Exterieur() bool { return o.Cotes[0] == RDehors || o.Cotes[1] == RDehors }

// Autre : la région de l'autre côté.
func (o *Obstacle) Autre(reg int) int {
	if o.Cotes[0] == reg {
		return o.Cotes[1]
	}
	return o.Cotes[0]
}

// Pièces : les noms donnés aux treize pièces de la villa (repérées par une
// de leurs cases, la numérotation venant du relevé).
var nomsPieces = []struct {
	c, r int
	nom  string
}{
	{3, 8, "le jardin d'hiver"},
	{8, 8, "la salle à manger"},
	{14, 8, "le grand salon"},
	{20, 8, "la bibliothèque"},
	{2, 13, "le boudoir"},
	{6, 13, "le salon de musique"},
	{12, 15, "le hall"},
	{19, 14, "la galerie"},
	{4, 19, "le fumoir"},
	{14, 19, "l'office"},
	{17, 19, "la chambre de Monsieur"},
	{8, 19, "le vestiaire"},
	{21, 19, "la chambre de Madame"},
}

// NomPiece : le nom d'une région.
func NomPiece(reg int) string {
	if reg == RDehors {
		return "le jardin"
	}
	for _, n := range nomsPieces {
		if P.Piece[Indice(n.c, n.r)] == reg {
			return n.nom
		}
	}
	return "une pièce"
}

// NomLieu : la région d'une case, en clair.
func NomLieu(k int) string {
	switch P.Terrain[k] {
	case TEnceinte:
		return "le mur d'enceinte"
	case TGrille:
		return "la grille"
	case TSeuil:
		return NomsObstacles[P.Obstacles[P.ObsCase[k]].Type]
	}
	return NomPiece(P.Region(k))
}

// Du : « de » contracté devant un nom de lieu (le hall → du hall).
func Du(s string) string {
	switch {
	case len(s) > 3 && s[:3] == "le ":
		return "du " + s[3:]
	case len(s) > 4 && s[:4] == "les ":
		return "des " + s[4:]
	}
	return "de " + s
}
