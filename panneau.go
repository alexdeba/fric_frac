package main

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"fricfrac/jeu"
)

// Le panneau de droite : en-tête Art déco, la consigne (une carte de
// visite crème, avec les actions possibles), les voleurs, les derniers dés
// et les « faits divers » du journal.

const (
	pX  = PanneauX + 18
	pW  = LargPanneau - 36
	pY0 = 96.0  // la carte de consigne
	pH0 = 318.0 // sa hauteur minimale (elle s'allonge s'il y a beaucoup d'actions)
	vH  = 44.0  // une ligne de voleur
)

func (g *Game) dessinerPanneau(dst *ebiten.Image) {
	rect(dst, PanneauX, 0, LargPanneau, HauteurEcran, coulNuit)
	rect(dst, PanneauX, 0, 3, HauteurEcran, coulOrSombre)
	cadreDeco(dst, PanneauX+8, 8, LargPanneau-16, HauteurEcran-16, coulOrSombre)
	// en-tête
	ecrireCentre(dst, "FRIC-FRAC", PanneauX+LargPanneau/2, 14, face(42, policeTitre), coulOr)
	e := g.p.E
	sous := fmt.Sprintf("Neuilly, la nuit · tour %d", max(e.Tour, 1))
	if e.Alarme {
		sous += " · ALERTE"
	}
	ecrireCentre(dst, sous, PanneauX+LargPanneau/2, 62, face(17, policeSans), coulCreme2)
	filetDeco(dst, pX+10, pX+pW-10, 88, coulOrSombre)
	g.dessinerConsigne(dst)
	g.dessinerListeVoleurs(dst)
	g.dessinerJournal(dst)
}

// carte : la carte de visite de la consigne.
func carteVisite(dst *ebiten.Image, x, y, w, h float64) {
	arrondi(dst, x+4, y+5, w, h, 6, alpha(coulNoir, 0.6), nil, 0)
	arrondi(dst, x, y, w, h, 6, coulCreme, coulOrSombre, 2)
	cadre(dst, x+7, y+7, w-14, h-14, 1, alpha(coulAcajou2, 0.6))
}

// consigne : titre et texte de l'étape en cours.
func (g *Game) consigne() (string, string) {
	e := g.p.E
	v := e.Voleurs[0]
	if g.finAffichee() {
		return "Passage chez le receleur", "La nuit est finie. Voici les comptes de chacun."
	}
	switch g.p.Req.Type {
	case ReqEquiper:
		return "Votre équipement", "Répartissez vos trois points d'équipement (sur le plateau), puis validez."
	case ReqAgir:
		o := g.p.Req.Options
		switch o.Phase {
		case jeu.PhaseEntree:
			return "Entrer dans la propriété", "Cliquez une case du mur d'enceinte (en doré) : au prochain tour vous sauterez dans le jardin. Ou postez-vous devant la grille (en vert) pour la forcer dès maintenant."
		case jeu.PhaseEffraction:
			return "Forcer", "Plus vous restez de tours sur place, moins vous faites de bruit. Un garde ou un dormeur à portée du bruit donne l'alarme."
		case jeu.PhaseMarche:
			t := fmt.Sprintf("Il vous reste %d point%s de déplacement. Cliquez une case surlignée.", o.PM, pl(o.PM))
			if !v.Equip.Chaussures {
				t += " Vos pas s'entendent à autant de cases que vous en parcourez (cercle)."
			}
			switch {
			case o.PeutSortir && len(o.Sauts) > 0:
				t = "Vous êtes sur le mur. Quittez la propriété, ou sautez dans le jardin : cliquez une case rouge (jet de saut)."
			case o.PeutSortir:
				t = "Vous êtes sur le mur. Quittez la propriété, ou restez ici."
			case len(o.Sauts) > 0:
				t = "Sautez du mur dans le jardin : cliquez une case rouge (jet de saut)."
			}
			return "Se déplacer", t
		case jeu.PhaseCombat:
			return "Combattre ?", "Un adversaire est à portée, dans la même pièce. Attaquez-le (choisissez l'arme) ou passez."
		}
	}
	if !v.Actif() {
		switch v.Statut {
		case jeu.Sorti:
			return "Vous êtes sorti", "Vos rivaux finissent leur nuit… (A : accélérer)"
		case jeu.Arrete:
			return "Arrêté !", "Les gardes vous ont mis la main au collet. Vos rivaux continuent… (A : accélérer)"
		case jeu.Abattu:
			return "Abattu", "Votre carrière s'arrête ici. Vos rivaux continuent… (A : accélérer)"
		}
	}
	if e.Courant >= 0 && e.Courant < len(e.Voleurs) && e.Courant != 0 {
		return "Au tour de " + e.Voleurs[e.Courant].Nom, "Les autres cambrioleurs jouent, puis déplacent les gardes."
	}
	return "Patientez", "Les gardes font leur ronde…"
}

func pl(n int) string {
	if n > 1 {
		return "s"
	}
	return ""
}

func (g *Game) dessinerConsigne(dst *ebiten.Image) {
	pH0 := g.hCarte
	carteVisite(dst, pX, pY0, pW, pH0)
	titre, texte := g.consigne()
	ecrireCentre(dst, titre, pX+pW/2, pY0+16, face(25, policeGras), coulBordeaux)
	paragraphe(dst, texte, pX+22, pY0+56, pW-44, face(18, policeTexte), coulAcajou, 24)
	// dés récents
	if len(g.p.Jets) > 0 && g.p.Req.Type != ReqEquiper {
		j := g.p.Jets[len(g.p.Jets)-1]
		y := pY0 + pH0 - 54
		x := pX + 22
		for _, d := range j.Des {
			de(dst, x, y, 34, d, coulCreme2, coulAcajou)
			x += 40
		}
		ecrire(dst, j.Titre, x+6, y-2, face(15, policeSansGras), coulBordeaux)
		t := j.Texte
		f := face(15, policeTexte)
		for largeur(t, f) > pX+pW-x-24 && len(t) > 4 {
			t = string([]rune(t)[:len([]rune(t))-2])
		}
		ecrire(dst, t, x+6, y+17, f, coulAcajou)
	}
}

// boutonsJeu : les actions de la demande en cours, la barre du bas.
func (g *Game) boutonsJeu() {
	g.hCarte = pH0
	by := float64(HauteurEcran) - 58
	bw := (pW - 20) / 3
	b := g.bouton(pX, by, bw, 40, "Règle", true, func() { g.regles = true })
	b.Style, b.Touche = BoutonDiscret, "R"
	b = g.bouton(pX+bw+10, by, bw, 40, "Zoom", true, func() { g.basculerZoom() })
	b.Style, b.Touche = BoutonDiscret, "Espace"
	txt := "Accélérer"
	if g.p.Rapide {
		txt = "Ralentir"
	}
	b = g.bouton(pX+2*bw+20, by, bw, 40, txt, true, func() { g.p.Rapide = !g.p.Rapide })
	b.Style, b.Touche = BoutonDiscret, "A"
	if g.finAffichee() {
		g.boutonsFin()
		return
	}
	switch g.p.Req.Type {
	case ReqEquiper:
		g.boutonsEquipement()
	case ReqAgir:
		g.boutonsAction()
	}
}

func (g *Game) boutonsAction() {
	o := g.p.Req.Options
	x, w := pX+22, pW-44
	_, texte := g.consigne()
	y := pY0 + 68 + float64(len(couper(texte, face(18, policeTexte), pW-44)))*24
	// le bouton « passer » sous les actions ; la carte s'allonge au besoin
	// (les dés s'affichent en bas de la carte)
	passer := func(t string) {
		yb := max(y+4, pY0+pH0-58-50)
		b := g.bouton(x, yb, w, 44, t, true, func() { g.p.repondre(jeu.Choix{Type: jeu.Passer}) })
		b.Touche = "Entrée"
		g.hCarte = max(pH0, yb+44+62-pY0)
	}
	switch o.Phase {
	case jeu.PhaseEffraction:
		for i, eff := range o.Effractions {
			if i >= 3 {
				break
			}
			// le nom au-dessus, puis les durées utiles sur toute la largeur
			nom := strings.ToUpper(jeu.NomsForcables[eff.Forcable][:1]) + jeu.NomsForcables[eff.Forcable][1:]
			ecrireDiffere(g, nom, x, y)
			y += 26
			durees := dureesUtiles(eff)
			bw := w / float64(len(durees))
			for j, n := range durees {
				i, n := i, n
				br := eff.Bruits[n-1]
				sous := fmt.Sprintf("bruit à %d", br)
				if br == 0 {
					sous = "silence"
				}
				b := g.bouton(x+float64(j)*bw, y, bw-6, 54, fmt.Sprintf("%d tour%s", n, pl(n)), true, func() {
					g.p.repondre(jeu.Choix{Type: jeu.CForcer, Indice: i, Tours: n})
				})
				b.Sous, b.Taille = sous, 20
				if br == 0 {
					b.Style = BoutonPrincipal
				}
			}
			y += 62
		}
		passer("Ne rien forcer")
	case jeu.PhaseMarche:
		if o.PeutSortir {
			b := g.bouton(x, y, w, 46, "Quitter la propriété", true, func() { g.p.repondre(jeu.Choix{Type: jeu.CSortir}) })
			b.Style = BoutonPrincipal
			y += 54
		}
		if o.Passage {
			b := g.bouton(x, y, w, 46, "Fuir par le passage secret", true, func() { g.p.repondre(jeu.Choix{Type: jeu.CPassage}) })
			b.Style = BoutonPrincipal
			y += 54
		}
		passer("Rester ici")
	case jeu.PhaseCombat:
		for i, a := range o.Adversaires {
			if i >= 3 {
				break
			}
			ecrireDiffere(g, nomAdversaire(g.p.E, a), x, y)
			y += 26
			bw := w / float64(len(a.Armes))
			for j, arme := range a.Armes {
				i, arme := i, arme
				t := map[jeu.Arme]string{jeu.AMainNue: "Poings", jeu.AMatraque: "Matraque", jeu.ACouteau: "Couteau", jeu.ARevolver: "Revolver"}[arme]
				if arme == jeu.ACouteau && a.Dist > 1 {
					t = "Lancer"
				}
				b := g.bouton(x+float64(j)*bw, y, bw-6, 46, t, true, func() {
					g.p.repondre(jeu.Choix{Type: jeu.CAttaquer, Indice: i, Arme: arme})
				})
				b.Taille = 20
			}
			y += 54
		}
		passer("Pas de combat")
	}
}

// dureesUtiles : les durées qui changent le bruit, jusqu'au silence
// (quatre au plus : 1, 2, 3 et le silence).
func dureesUtiles(eff jeu.Effraction) []int {
	var l []int
	for n := 1; n <= 8; n++ {
		if n > 1 && eff.Bruits[n-1] == eff.Bruits[n-2] {
			continue
		}
		l = append(l, n)
		if eff.Bruits[n-1] == 0 {
			break
		}
	}
	if len(l) > 4 {
		l = append(l[:3], l[len(l)-1])
	}
	return l
}

func nomAdversaire(e *jeu.Etat, a jeu.Adversaire) string {
	d := ""
	if a.Dist > 1 {
		d = fmt.Sprintf(" (à %d)", a.Dist)
	}
	switch a.Type {
	case jeu.AdvGarde:
		return jeu.Majuscule(e.Gardes[a.Indice].Nom) + d
	case jeu.AdvVoleur:
		return e.Voleurs[a.Indice].Nom + d
	}
	return "Le dormeur" + d
}

// ecrireDiffere : un libellé de ligne d'action, écrit au dessin (les
// boutons sont construits dans Update).
func ecrireDiffere(g *Game, s string, x, y float64) {
	g.libelles = append(g.libelles, libelle{s, x, y})
}

type libelle struct {
	s    string
	x, y float64
}

func (g *Game) dessinerLibelles(dst *ebiten.Image) {
	for _, l := range g.libelles {
		ecrire(dst, l.s, l.x, l.y, face(18, policeGras), coulAcajou)
	}
}

func (g *Game) dessinerListeVoleurs(dst *ebiten.Image) {
	g.dessinerLibelles(dst)
	e := g.p.E
	y := pY0 + g.hCarte + 14
	ecrire(dst, "LES CAMBRIOLEURS", pX+4, y, face(16, policeSansGras), coulOr)
	y += 24
	for _, v := range e.Voleurs {
		if e.Courant == v.Num && !g.finAffichee() {
			rect(dst, pX, y-2, pW, vH-2, alpha(coulOr, 0.14))
		}
		jeton(dst, imgVoleurs[v.Portrait], pX+22, y+vH/2-2, 38, ebiten.ColorScale{})
		cadreCarre(dst, pX+22, y+vH/2-2, 38, 2.5, coulVoleurs[v.Portrait])
		nom := v.Nom
		if v.Humain {
			nom += " (vous)"
		}
		ecrire(dst, nom, pX+50, y, face(18, policeGras), coulCreme)
		etat := ""
		switch v.Statut {
		case jeu.Attente:
			etat = "devant la propriété"
		case jeu.EnJeu:
			etat = jeu.NomLieu(v.Case)
			if v.Immobile > 0 {
				etat += fmt.Sprintf(" · sur place (%d)", v.Immobile)
			}
			if v.Travail != nil {
				etat += " · force " + jeu.NomsForcables[v.Travail.Eff.Forcable]
			}
		case jeu.Sorti:
			etat = "sorti · " + jeu.Francs(v.Score) + " chez le receleur"
		case jeu.Arrete:
			etat = "arrêté"
		case jeu.Abattu:
			etat = "abattu"
		}
		ecrire(dst, etat, pX+50, y+22, face(14, policeSans), coulCreme2)
		if v.Statut == jeu.EnJeu || v.Statut == jeu.Attente {
			val := jeu.Francs(v.Valeur())
			ecrireDroite(dst, val, pX+pW-4, y+2, face(18, policeSansGras), coulOr2)
			marques := ""
			if v.Vu {
				marques += "vu "
			}
			if v.ATue {
				marques += "· a tué"
			} else if v.AAssomme {
				marques += "· a assommé"
			}
			ecrireDroite(dst, marques, pX+pW-4, y+24, face(13, policeSans), coulRouge)
		}
		y += vH
	}
	g.yJournal = y + 8
}

func (g *Game) dessinerJournal(dst *ebiten.Image) {
	y0 := g.yJournal
	y1 := float64(HauteurEcran) - 70
	if y1-y0 < 60 {
		return
	}
	rect(dst, pX, y0, pW, y1-y0, coulCreme2)
	cadre(dst, pX, y0, pW, y1-y0, 2, coulOrSombre)
	ecrireCentre(dst, "— FAITS DIVERS —", pX+pW/2, y0+6, face(16, policeSansGras), coulAcajou)
	f := face(15, policeMachine)
	inter := 19.0
	// les lignes, de la plus récente vers le haut
	var lignes []struct {
		t    string
		fort bool
		j    int
	}
	for _, l := range g.p.Journal {
		for _, s := range couper(l.Texte, f, pW-24) {
			lignes = append(lignes, struct {
				t    string
				fort bool
				j    int
			}{s, l.Fort, l.Joueur})
		}
	}
	n := int((y1 - y0 - 34) / inter)
	fin := max(len(lignes)-g.defil, 0)
	debut := max(fin-n, 0)
	y := y0 + 30
	for _, l := range lignes[debut:fin] {
		c := coulAcajou
		if l.fort {
			c = coulBordeaux
		}
		if l.j == 0 {
			rect(dst, pX+5, y+3, 3, inter-6, coulOrSombre)
		}
		ecrire(dst, l.t, pX+12, y, f, c)
		y += inter
	}
}

// ------------------------------------------------------------ équipement

type article struct {
	nom   string
	cout  int
	descr string
	a     func(q *jeu.Equipement) *bool // pour les cases à cocher
	niv   func(q *jeu.Equipement) *int  // pour les niveaux
	n     int
}

var articles = []article{
	{nom: "Passe-partout, 1 tour", cout: 2, descr: "Compte comme un tour de plus pour forcer une serrure.", niv: func(q *jeu.Equipement) *int { return &q.Passe }, n: 1},
	{nom: "Passe-partout, 2 tours", cout: 3, descr: "Deux tours de plus sur la table d'effraction.", niv: func(q *jeu.Equipement) *int { return &q.Passe }, n: 2},
	{nom: "Chaussures silencieuses", cout: 1, descr: "Vos déplacements ne font aucun bruit.", a: func(q *jeu.Equipement) *bool { return &q.Chaussures }},
	{nom: "Pare-balles, 1", cout: 1, descr: "Ôte 1 point au dé de l'adversaire.", niv: func(q *jeu.Equipement) *int { return &q.PareBalles }, n: 1},
	{nom: "Pare-balles, 2", cout: 2, descr: "Ôte 2 points au dé de l'adversaire.", niv: func(q *jeu.Equipement) *int { return &q.PareBalles }, n: 2},
	{nom: "Matraque", cout: 1, descr: "Assomme plus sûrement qu'à mains nues.", a: func(q *jeu.Equipement) *bool { return &q.Matraque }},
	{nom: "Couteau", cout: 2, descr: "Au contact ou lancé (un point de moins par case).", a: func(q *jeu.Equipement) *bool { return &q.Couteau }},
	{nom: "Revolver", cout: 3, descr: "Redoutable, mais s'entend dans toute la villa.", a: func(q *jeu.Equipement) *bool { return &q.Revolver }},
}

func (art article) choisi(q jeu.Equipement) bool {
	if art.a != nil {
		return *art.a(&q)
	}
	return *art.niv(&q) == art.n
}

// basculer : coche ou décoche l'article (les niveaux s'excluent).
func (art article) basculer(q jeu.Equipement) jeu.Equipement {
	if art.a != nil {
		p := art.a(&q)
		*p = !*p
		return q
	}
	p := art.niv(&q)
	if *p == art.n {
		*p = 0
	} else {
		*p = art.n
	}
	return q
}

const (
	eqX, eqY, eqW, eqH = 170.0, 120.0, 760.0, 740.0
)

func (g *Game) dessinerEquipement(dst *ebiten.Image) {
	rect(dst, 0, 0, LargVue, HauteurEcran, alpha(coulNoir, 0.55))
	carteVisite(dst, eqX, eqY, eqW, eqH)
	ecrireCentre(dst, "L'ÉQUIPEMENT DU CAMBRIOLEUR", eqX+eqW/2, eqY+26, face(30, policeTitre), coulBordeaux)
	reste := jeu.PointsEquipement - g.equip.Cout()
	filetDeco(dst, eqX+160, eqX+eqW-160, eqY+80, coulOrSombre)
	ecrireCentre(dst, fmt.Sprintf("Points restants : %d sur 3", reste), eqX+eqW/2, eqY+96, face(22, policeGras), coulAcajou)
	for i, art := range articles {
		y := eqY + 150 + float64(i)*62
		ecrire(dst, art.descr, eqX+330, y+12, face(16, policeItalique), coulAcajou)
	}
}

func (g *Game) boutonsEquipement() {
	for i, art := range articles {
		art := art
		y := eqY + 150 + float64(i)*62
		suivant := art.basculer(g.equip)
		ok := art.choisi(g.equip) || suivant.Cout() <= jeu.PointsEquipement
		b := g.bouton(eqX+40, y, 270, 48, art.nom, ok, func() { g.equip = art.basculer(g.equip) })
		b.Sous = fmt.Sprintf("%d point%s", art.cout, pl(art.cout))
		b.Taille = 18
		b.Choisi = art.choisi(g.equip)
	}
	b := g.bouton(eqX+eqW/2-160, eqY+eqH-70, 320, 50, "Valider l'équipement", g.equip.Valide(), func() {
		g.p.repondre(g.equip)
	})
	b.Style = BoutonPrincipal
}

// ------------------------------------------------------------ fin

func (g *Game) dessinerFin(dst *ebiten.Image) {
	e := g.p.E
	x, y, w, h := 130.0, 70.0, 840.0, 800.0
	rect(dst, 0, 0, LargVue, HauteurEcran, alpha(coulNoir, 0.55))
	carteVisite(dst, x, y, w, h)
	ecrireCentre(dst, "CHEZ LE RECELEUR", x+w/2, y+28, face(36, policeTitre), coulBordeaux)
	filetDeco(dst, x+120, x+w-120, y+84, coulOrSombre)
	best := -1
	for _, v := range e.Voleurs {
		if v.Statut == jeu.Sorti && (best < 0 || v.Score > e.Voleurs[best].Score) {
			best = v.Num
		}
	}
	yy := y + 110
	for _, v := range e.Voleurs {
		jeton(dst, imgVoleurs[v.Portrait], x+70, yy+40, 70, ebiten.ColorScale{})
		cadreCarre(dst, x+70, yy+40, 70, 3, coulVoleurs[v.Portrait])
		nom := v.Nom
		if v.Humain {
			nom += " (vous)"
		}
		ecrire(dst, nom, x+124, yy+6, face(24, policeGras), coulAcajou)
		var detail string
		switch v.Statut {
		case jeu.Sorti:
			detail = fmt.Sprintf("Sorti au tour %d avec %s de butin", v.TourFin, jeu.Francs(v.Valeur()))
			if d := v.Diviseur(); d > 1 {
				detail += fmt.Sprintf(" — objets divisés par %d", d)
			}
		case jeu.Arrete:
			detail = fmt.Sprintf("Arrêté au tour %d : tout est perdu", v.TourFin)
		case jeu.Abattu:
			detail = fmt.Sprintf("Abattu au tour %d", v.TourFin)
		}
		ecrire(dst, detail, x+124, yy+42, face(17, policeTexte), coulAcajou2)
		score := 0
		if v.Statut == jeu.Sorti {
			score = v.Score
		}
		ecrireDroite(dst, jeu.Francs(score), x+w-40, yy+14, face(28, policeSansGras), coulBordeaux)
		if v.Num == best {
			ecrireDroite(dst, "LE PLUS RICHE", x+w-40, yy+50, face(15, policeSansGras), coulOrSombre)
		}
		yy += 84
	}
}

func (g *Game) boutonsFin() {
	b := g.bouton(300, 790, 240, 50, "Nouvelle partie", true, func() {
		g.p = nil
		g.ecran = EcranAccueil
	})
	b.Style = BoutonPrincipal
	b = g.bouton(560, 790, 240, 50, "Quitter", true, func() { quitter() })
	b.Style = BoutonDiscret
}
