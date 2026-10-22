package jeu

import (
	"os"
	"testing"
	"time"
)

// BOUCLE=1 : des parties à 6 sur beaucoup de graines, chacune bornée dans
// le temps (détection de boucle infinie).
func TestBoucle(t *testing.T) {
	if os.Getenv("BOUCLE") == "" {
		t.Skip()
	}
	for g := range uint64(300) {
		fini := make(chan bool)
		go func() { partieIA(6, g); fini <- true }()
		select {
		case <-fini:
		case <-time.After(5 * time.Second):
			t.Fatalf("graine %d : la partie ne finit pas", g)
		}
	}
}
