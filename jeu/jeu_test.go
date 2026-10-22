package jeu

import (
	"fmt"
	"os"
	"testing"
)

// Le plateau relevé : treize pièces, les bases des gardes dans le hall,
// les chiffres autour des bases sur des cases du hall.
func TestPlateau(t *testing.T) {
	if NbPieces != 13 {
		t.Fatalf("%d pièces, 13 attendues", NbPieces)
	}
	hall := P.Piece[P.BaseA]
	if hall < 0 || P.Piece[P.BaseB] != hall || NomPiece(hall) != "le hall" {
		t.Fatalf("bases hors du hall : %d %d", P.Piece[P.BaseA], P.Piece[P.BaseB])
	}
	for f := range 6 {
		for _, b := range []struct {
			base int
			dirs [6]int
		}{{P.BaseA, P.DirsA}, {P.BaseB, P.DirsB}} {
			if v := Voisin(b.base, b.dirs[f]); P.Piece[v] != hall {
				t.Errorf("face %d : la case voisine n'est pas dans le hall", f+1)
			}
		}
	}
	// chaque pièce a une porte ou une fenêtre
	acces := map[int]bool{}
	for _, o := range P.Obstacles {
		for _, c := range o.Cotes {
			acces[c] = true
		}
	}
	for p := range NbPieces {
		if !acces[p] {
			t.Errorf("%s sans porte ni fenêtre", NomPiece(p))
		}
	}
	nomsVus := map[string]bool{}
	for p := range NbPieces {
		n := NomPiece(p)
		if nomsVus[n] || n == "une pièce" {
			t.Errorf("pièce %d mal nommée : %s", p, n)
		}
		nomsVus[n] = true
	}
	if len(P.Grilles) != 3 {
		t.Errorf("%d cases de grille", len(P.Grilles))
	}
}

func TestDist(t *testing.T) {
	a := Indice(5, 5)
	for d := range 6 {
		if Dist(a, Voisin(a, d)) != 1 {
			t.Errorf("voisin %d à distance %d", d, Dist(a, Voisin(a, d)))
		}
		b := Voisin(Voisin(a, d), d)
		if Dist(a, b) != 2 {
			t.Errorf("deux pas %d : distance %d", d, Dist(a, b))
		}
	}
}

func TestTables(t *testing.T) {
	if Bruit(FCoffre, 1, 0) != 7 || Bruit(FCoffre, 8, 0) != 0 || Bruit(FPorteInt, 1, 1) != 1 || Bruit(FFenetre, 1, 2) != 1 {
		t.Error("table d'effraction")
	}
	if ResultatCombat(ARevolver, -3) != KO || ResultatCombat(ACouteau, 5) != Tue || ResultatCombat(ASaut, 4) != KO {
		t.Error("table de combat")
	}
	if b := TableButin(66); b.Objet.Valeur != 500000 {
		t.Error("rivière de diamants")
	}
	n := 0
	for d1 := 1; d1 <= 6; d1++ {
		for d2 := 1; d2 <= 6; d2++ {
			if TableButin(10*d1+d2).Type == BObjet {
				n++
			}
		}
	}
	if n != 21 {
		t.Errorf("%d cases à butin, 21 attendues", n)
	}
	q := Equipement{Passe: 1, Chaussures: true}
	if !q.Valide() || (Equipement{Revolver: true, Matraque: true}).Valide() {
		t.Error("équipement")
	}
}

// Meubles : 36 pions, aucun contre une porte, jamais deux adjacents.
func TestMeubles(t *testing.T) {
	for g := range uint64(30) {
		e := NouvelEtat([]string{"a", "b"}, []int{0, 1}, []bool{false, false}, g)
		if len(e.Meubles) != 36 {
			t.Fatalf("%d meubles", len(e.Meubles))
		}
		for i, m := range e.Meubles {
			for _, c := range m.Cases {
				if P.Terrain[c] != TPiece {
					t.Fatalf("meuble hors d'une pièce")
				}
				for j, n := range e.Meubles {
					if i != j && n.Adjacent(c) {
						t.Fatalf("meubles adjacents")
					}
				}
			}
		}
	}
}

// partie entre ordinateurs, jusqu'au bout
func partieIA(n int, graine uint64) *Etat {
	noms := make([]string, n)
	portraits := make([]int, n)
	hum := make([]bool, n)
	for i := range n {
		noms[i] = NomsVoleurs[i]
		portraits[i] = i
	}
	e := NouvelEtat(noms, portraits, hum, graine)
	for range n {
		e.Ctrl = append(e.Ctrl, IA{})
	}
	e.Jouer()
	return e
}

type journal struct{ lignes []string }

func (j *journal) Evenement(e *Etat, ev Ev) {
	if ev.Texte != "" && ev.Type != EvDes {
		j.lignes = append(j.lignes, fmt.Sprintf("T%d %s", e.Tour, ev.Texte))
	}
}

func TestParties(t *testing.T) {
	sortis, arretes, abattus, total, tours := 0, 0, 0, 0, 0
	gain := 0
	for g := range uint64(60) {
		e := partieIA(1+int(g%6), g)
		for _, v := range e.Voleurs {
			total++
			switch v.Statut {
			case Sorti:
				sortis++
				gain += v.Score
			case Arrete:
				arretes++
			case Abattu:
				abattus++
			default:
				t.Fatalf("graine %d : %s encore en jeu (%d)", g, v.Nom, v.Statut)
			}
			tours += v.TourFin
		}
	}
	t.Logf("%d voleurs : %d sortis (gain moyen %s), %d arrêtés, %d abattus ; %d tours en moyenne",
		total, sortis, Francs(gain/max(sortis, 1)), arretes, abattus, tours/total)
	if sortis == 0 {
		t.Error("aucun voleur ne ressort")
	}
}

// PARTIE=graine : le journal complet d'une partie à 3.
func TestJournal(t *testing.T) {
	var g uint64
	if _, err := fmt.Sscan(os.Getenv("PARTIE"), &g); err != nil {
		t.Skip("PARTIE non fourni")
	}
	noms := []string{NomsVoleurs[0], NomsVoleurs[1], NomsVoleurs[2]}
	e := NouvelEtat(noms, []int{0, 1, 2}, []bool{false, false, false}, g)
	e.Ctrl = []Controleur{IA{}, IA{}, IA{}}
	j := &journal{}
	e.Obs = j
	e.Jouer()
	for _, l := range j.lignes {
		fmt.Println(l)
	}
}
