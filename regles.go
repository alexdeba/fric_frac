package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"fricfrac/jeu"
)

// La règle, consultable en jeu (touche R) : le texte du magazine, les
// trois tables, et les choix faits pour l'adaptation.

type pageRegle struct {
	titre  string
	textes []string // paragraphes ; « # » en tête : intertitre
	table  int      // 0 aucune, 1 effraction, 2 combat, 3 butin
}

var pagesRegle = []pageRegle{
	{titre: "Le casse de votre vie", textes: []string{
		"Cambrioleur, mais néanmoins gentleman, vous surveillez nuit et jour depuis deux mois ce petit hôtel particulier de Neuilly : les allées et venues de ses richissimes propriétaires, la ronde des deux gardes… Vous venez d'apprendre au hasard d'une conversation que Monsieur vient d'offrir à Madame une rivière de diamants que l'on estime à… Animé par la passion du beau d'une part, l'attrait du danger d'autre part, vous avez pris votre décision : le « casse » de votre vie est pour ce soir…",
		"# But du jeu",
		"Pénétrer dans la villa pour rafler, sans se faire voir des gardes, le plus grand nombre d'objets de valeur, qui se trouvent dans le mobilier.",
		"# Mise en place",
		"Les 36 pions-mobilier sont installés au hasard, face cachée, dans les pièces de la villa. On ne doit ni obstruer une porte, ni placer deux meubles sur des cases adjacentes. Les deux gardes partent de leurs bases, A et B, dans le hall.",
		"# Équipement",
		"Chaque joueur dispose de trois points pour équiper son voleur : passe-partout 1 tour (2 pts) ou 2 tours (3 pts) pour ouvrir plus vite les serrures ; chaussures silencieuses (1 pt) ; pare-balles 1 (1 pt) ou 2 (2 pts) ; matraque (1 pt), couteau (2 pts), revolver (3 pts).",
	}},
	{titre: "Déroulement de la partie", textes: []string{
		"À chaque tour, les joueurs s'efforcent d'ouvrir silencieusement les serrures qui se présentent à eux, de rafler éventuellement un butin. Ils déplacent ensuite leur pion voleur, toujours le plus silencieusement possible, effectuent un combat contre les gardes ou les autres voleurs s'ils l'estiment nécessaire, puis déplacent les gardes.",
		"# Entrer",
		"Au début, tous les voleurs sont hors du plateau. Pour pénétrer dans le jardin, on pose son pion sur une case-mur de son choix ; le tour suivant, on saute dans le jardin en consultant la table des combats (colonne « saut »). Les pare-balles ne protègent pas des conséquences du saut. On peut aussi ouvrir la grille dès le premier tour, avec la table d'effraction.",
		"# Effraction",
		"Les grilles, les portes intérieures et extérieures (rouges) et les fenêtres (bleues) sont des obstacles que le voleur doit franchir par effraction. Tous les meubles doivent être forcés pour s'emparer du butin. L'effraction doit être effectuée même si l'obstacle a déjà été franchi.",
		"On fait d'autant plus de bruit qu'on veut aller vite : chaque tour passé sur place diminue la portée du bruit (table d'effraction). Si un garde ou un dormeur peut l'entendre, l'alarme est donnée. Pour franchir une porte ou une fenêtre, il faut être, au début de son tour, sur la case où se trouve l'obstacle ; pour un meuble, sur une case adjacente. Plusieurs effractions sont possibles dans le même tour sans déplacer le pion.",
	}},
	{titre: "Le butin", textes: []string{
		"On ne trouve de butin que dans les secrétaires, les armoires et les coffres-forts. Une fois le meuble forcé, on lance deux dés : le premier donne les dizaines, le second les unités, et l'on consulte la table de butin ; le résultat s'applique immédiatement. On lance une fois pour un secrétaire, deux fois pour une armoire, trois fois pour un coffre-fort.",
		"Le passage secret permet de sortir du plateau sans effraction, en passant par le meuble en question ; seul le joueur qui l'a trouvé peut l'utiliser, au moment où il le désire.",
		"# Les meubles",
		"On ne peut pas passer sur une case occupée par un meuble. Dès que le pion d'un joueur passe sur une case adjacente à un meuble, on le retourne. Le joueur décide alors s'il continue son mouvement ou s'il reste en place pour forcer le meuble au tour suivant. Dès que le pion du joueur qui a découvert le meuble s'éloigne, celui-ci est considéré comme vide, qu'il ait été forcé ou non.",
		"Si le meuble découvert est un lit, il faut s'éloigner sans bruit : le dormeur se réveille et donne l'alarme si la case de sa tête est dans une zone de bruit.",
	}},
	{titre: "Les déplacements", textes: []string{
		"# Les voleurs",
		"Pendant leur tour, les voleurs se déplacent dans toutes les directions d'autant de cases qu'il y a de joueurs dans la partie (ou moins). À moins de porter des chaussures silencieuses, un déplacement d'une case s'entend à une case du point d'arrivée, de deux cases à deux cases, et ainsi de suite.",
		"Les escaliers se franchissent comme des cases normales, ainsi que les fenêtres près d'un arbre ; les autres fenêtres coûtent deux points de déplacement pour entrer dans la villa, et si l'on en sort, on saute (table des combats).",
		"# Les gardes",
		"Dès qu'un voleur entre dans la villa, les gardes commencent leur ronde. Un dé donne la direction de chacun (les chiffres autour de sa base) ; ils vont en ligne droite jusqu'à un obstacle (mur, meuble, fenêtre ou porte extérieure), puis on relance le dé jusqu'à trouver une direction libre. Chaque joueur déplace les deux gardes d'une case.",
		"Si l'alarme est donnée, s'ils sont dans la zone d'un bruit ou dans la même pièce qu'un voleur, les gardes se déplacent de deux cases par tour jusqu'à la fin de la partie, par le plus court chemin vers la case suspecte ; s'il n'y a personne, ils reprennent leur ronde. Un garde qui voit un voleur le poursuit jusqu'à son arrestation, même dans le jardin.",
	}},
	{titre: "Les combats", textes: []string{
		"Deux pions de la même pièce peuvent se battre. À mains nues, à la matraque ou au couteau, il faut être sur des cases adjacentes ; au revolver ou au couteau lancé, chaque case qui sépare les adversaires ôte un point au dé.",
		"Entre voleurs, celui dont c'est le tour attaque ; si l'adversaire n'est pas neutralisé, il riposte, et ainsi de suite. Le vainqueur peut s'emparer du butin de l'autre. Contre un dormeur, le voleur doit frapper avant que celui-ci ne donne l'alarme : si le résultat est OK, le dormeur donne aussitôt l'alarme.",
		"Contre un garde, c'est toujours le voleur qui attaque. Le garde veut arrêter le voleur en se plaçant sur sa case ; s'il y parvient, la partie est finie pour ce voleur. Les gardes sont armés de revolvers.",
		"OK : le défenseur est indemne. KO : assommé ou blessé, il reste sur place jusqu'à ce que l'attaquant ait joué deux tours (deux tours sur place après un saut raté). Tué : si un garde entre dans la pièce où se trouve le cadavre, l'alarme est donnée. Un coup de revolver s'entend dans toute la villa et donne l'alarme.",
		"# Fin de partie",
		"Quand un joueur estime avoir assez de butin, il fait sortir son pion. Chez le receleur, l'argent des portefeuilles, les pièces et les lingots d'or comptent intégralement. Le reste vaut la valeur de la table si le voleur n'a pas été vu ; elle est divisée par cinq s'il a assommé un garde ou un dormeur, par dix s'il en a tué un. Le plus riche gagne.",
	}},
	{titre: "Table d'effraction et table de combat", table: 1},
	{titre: "Table de butin", table: 3},
	{titre: "Choix d'adaptation", textes: []string{
		"Le plateau est la photo de l'encart, redressée ; les cases, murs, portes et fenêtres ont été relevés dessus. Une porte ou une fenêtre peut couper une case (on s'y place pour la forcer) ou passer entre deux cases (on la force depuis l'une ou l'autre).",
		"Les meubles sont toujours placés au hasard par l'ordinateur. Les armoires et les lits tiennent deux cases ; la tête du dormeur est du côté de l'oreiller. Un dormeur dort même sous un pion encore caché.",
		"Le passe-partout compte comme un ou deux tours de plus pour toute serrure (grille, portes, meubles), pas pour les fenêtres. Le passage secret se prend depuis une case voisine du meuble.",
		"Depuis une case-mur, on saute dans une case voisine du jardin et le déplacement s'arrête là ; on quitte la propriété depuis n'importe quelle case-mur. Les gardes ne montent pas sur le mur et ne passent pas par les fenêtres.",
		"Assommé : deux tours. Un dormeur assommé se réveille ensuite et donne l'alarme. Un voleur vu sans avoir frappé personne voit aussi ses objets divisés par cinq. Au lever du jour (tour 150), ceux qui sont encore dans la propriété sont pris.",
	}},
}

func (g *Game) touchesRegle() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.regles = false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		g.pageRgl = min(g.pageRgl+1, len(pagesRegle)-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		g.pageRgl = max(g.pageRgl-1, 0)
	}
}

const (
	rgX, rgY, rgW, rgH = 180.0, 50.0, 1240.0, 900.0
)

func (g *Game) dessinerRegles(dst *ebiten.Image) {
	rect(dst, 0, 0, LargeurEcran, HauteurEcran, alpha(coulNoir, 0.7))
	carteVisite(dst, rgX, rgY, rgW, rgH)
	pg := pagesRegle[g.pageRgl]
	ecrireCentre(dst, "RÈGLE DU JEU", rgX+rgW/2, rgY+22, face(20, policeTitre), coulOrSombre)
	ecrireCentre(dst, pg.titre, rgX+rgW/2, rgY+50, face(34, policeGras), coulBordeaux)
	filetDeco(dst, rgX+300, rgX+rgW-300, rgY+104, coulOrSombre)
	ecrireCentre(dst, fmt.Sprintf("page %d / %d · flèches gauche et droite pour tourner les pages", g.pageRgl+1, len(pagesRegle)), rgX+rgW/2, rgY+rgH-48, face(17, policeSans), coulAcajou2)
	switch pg.table {
	case 1:
		g.tableEffraction(dst, rgX+70, rgY+130)
		g.tableCombat(dst, rgX+70, rgY+470)
		return
	case 3:
		g.tableButin(dst, rgX+120, rgY+130)
		return
	}
	y := rgY + 126
	for _, t := range pg.textes {
		if t[0] == '#' {
			y += 6
			ecrire(dst, t[2:], rgX+70, y, face(24, policeGras), coulBordeaux)
			y += 36
			continue
		}
		y += paragraphe(dst, t, rgX+70, y, rgW-140, face(20, policeTexte), coulAcajou, 28) + 10
	}
}

func (g *Game) tableEffraction(dst *ebiten.Image, x, y float64) {
	ecrire(dst, "Tableau 1 : table d'effraction — portée du bruit (cases) selon les tours sur place", x, y, face(19, policeGras), coulBordeaux)
	y += 38
	f := face(19, policeTexte)
	for n := 1; n <= 8; n++ {
		ecrireCentre(dst, fmt.Sprint(n), x+330+float64(n-1)*80, y, face(19, policeSansGras), coulAcajou)
	}
	lignes := []jeu.Forcable{jeu.FGrille, jeu.FPorteExt, jeu.FFenetre, jeu.FPorteInt, jeu.FSecretaire, jeu.FArmoire, jeu.FCoffre}
	noms := []string{"Grille", "Porte extérieure", "Fenêtre", "Porte intérieure", "Secrétaire", "Armoire", "Coffre-fort"}
	for i, l := range lignes {
		yy := y + 32 + float64(i)*34
		if i%2 == 0 {
			rect(dst, x-10, yy-4, 1000, 32, alpha(coulCreme2, 0.8))
		}
		ecrire(dst, noms[i], x, yy, f, coulAcajou)
		for n := 1; n <= 8; n++ {
			b := jeu.Bruit(l, n, 0)
			s := fmt.Sprint(b)
			c := coulBordeaux
			if b == 0 {
				s, c = "0", coulVert
			}
			ecrireCentre(dst, s, x+330+float64(n-1)*80, yy, f, c)
		}
	}
}

func (g *Game) tableCombat(dst *ebiten.Image, x, y float64) {
	ecrire(dst, "Tableau 2 : table de combat — issue pour le défenseur selon le dé", x, y, face(19, policeGras), coulBordeaux)
	y += 38
	cols := []string{"saut", "main nue", "matraque", "couteau", "revolver"}
	for i, c := range cols {
		ecrireCentre(dst, c, x+250+float64(i)*150, y, face(19, policeSansGras), coulAcajou)
	}
	for d := 1; d <= 6; d++ {
		yy := y + 32 + float64(d-1)*34
		if d%2 == 1 {
			rect(dst, x-10, yy-4, 1000, 32, alpha(coulCreme2, 0.8))
		}
		ecrireCentre(dst, fmt.Sprint(d), x+60, yy, face(19, policeSansGras), coulAcajou)
		for i := range cols {
			r := jeu.ResultatCombat(jeu.Arme(i), d)
			s := []string{"OK", "KO", "tué"}[r]
			c := []any{coulVert, coulAcajou, coulBordeaux}[r]
			ecrireCentre(dst, s, x+250+float64(i)*150, yy, face(19, policeTexte), c.(interface{ RGBA() (r, g, b, a uint32) }))
		}
	}
	ecrire(dst, "Couteau lancé et revolver : un point de moins au dé par case d'écart. Pare-balles : 1 ou 2 points de moins au dé de l'adversaire.", x, y+250, face(17, policeItalique), coulAcajou2)
}

func (g *Game) tableButin(dst *ebiten.Image, x, y float64) {
	ecrire(dst, "Tableau 3 : deux dés, le premier pour les dizaines, le second pour les unités", x, y, face(19, policeGras), coulBordeaux)
	y += 44
	f := face(18, policeTexte)
	i := 0
	for d1 := 1; d1 <= 6; d1++ {
		for d2 := 1; d2 <= 6; d2++ {
			n := 10*d1 + d2
			b := jeu.TableButin(n)
			col := i / 18
			row := i % 18
			xx := x + float64(col)*520
			yy := y + float64(row)*34
			if row%2 == 0 {
				rect(dst, xx-8, yy-4, 500, 32, alpha(coulCreme2, 0.8))
			}
			ecrire(dst, fmt.Sprint(n), xx, yy, face(18, policeSansGras), coulBordeaux)
			t := b.Texte
			c := coulAcajou
			if b.Type == jeu.BObjet {
				t = fmt.Sprintf("%s : %s", b.Objet.Nom, jeu.Francs(b.Objet.Valeur))
				c = coulAcajou2
			}
			ecrire(dst, t, xx+46, yy, f, c)
			i++
		}
	}
}

func (g *Game) boutonsRegles() {
	b := g.bouton(rgX+40, rgY+rgH-62, 180, 44, "Précédente", g.pageRgl > 0, func() { g.pageRgl-- })
	b.Style = BoutonDiscret
	b = g.bouton(rgX+rgW-220, rgY+rgH-62, 180, 44, "Suivante", g.pageRgl < len(pagesRegle)-1, func() { g.pageRgl++ })
	b.Style = BoutonDiscret
	b = g.bouton(rgX+rgW/2-110, rgY+rgH-110, 220, 44, "Fermer", true, func() { g.regles = false })
	b.Style, b.Touche = BoutonPrincipal, "Échap"
}
