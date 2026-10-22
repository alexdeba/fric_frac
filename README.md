# Fric-Frac

D'après « fric-frac », jeu en encart de **François Nédélec** paru dans
*Jeux & Stratégie*. Cambrioleur
gentleman, vous visez un petit hôtel particulier de Neuilly gardé par deux
gardes ; jusqu'à cinq rivaux menés par l'ordinateur font la même nuit le
même casse.

## Compiler

Il faut [Go](https://go.dev/dl/) 1.27.1 ou plus récent. Le plateau, les
pions et les polices sont intégrés au binaire : il se suffit à lui-même.

```
git clone https://github.com/alexdeba/fric_frac.git
cd fric_frac
go build -o fricfrac .
./fricfrac                        # F11 : plein écran
```

`go run .` lance le jeu sans garder de binaire.

Selon le système :

- **Linux** : Ebitengine a besoin d'un compilateur C et des en-têtes OpenGL
  et X11. Sous Debian ou Ubuntu :
  ```
  sudo apt install gcc pkg-config libgl1-mesa-dev libxcursor-dev libxi-dev \
      libxinerama-dev libxrandr-dev libxxf86vm-dev
  ```
- **macOS** : les outils en ligne de commande de Xcode
  (`xcode-select --install`).
- **Windows** : rien de plus. `go build -o fricfrac.exe .` donne le
  binaire. On peut aussi le produire depuis Linux ou macOS :
  `GOOS=windows go build -o fricfrac.exe .`

## Commandes

- Accueil : nombre d'adversaires (0 à 5), choix de votre pion, Entrée pour
  commencer. Puis l'écran d'équipement (trois points à répartir).
- Le panneau de droite donne la consigne de l'étape en cours et ses
  actions ; sur le plateau, les coups possibles sont surlignés : cases-mur
  d'entrée (or, la grille en vert), cases atteignables (or ; rouge : saut
  par une fenêtre ou depuis le mur), obstacles et meubles à forcer (anneau
  rouge), adversaires à portée (anneau rouge).
- Survol d'une case : le lieu, le meuble ou le pion qui s'y trouve, le coût
  du trajet ; le chemin est tracé, avec la portée du bruit de vos pas si
  vous n'avez pas de chaussures silencieuses.
- Barre d'espace : zoom sur la souris (ou sur votre pion), et retour à tout
  le plateau. Molette : zoom ; clic droit glissé ou flèches : déplacer la
  vue ; `0` : tout le plateau.
- Plateau zoomé : une mini-carte en haut à droite montre toute la propriété,
  les pions et le cadre de la vue ; cliquer ou glisser dedans déplace la vue.
- `R` : la règle (texte du magazine, les trois tables, choix d'adaptation).
- `A` : accélérer l'ordinateur. Entrée ou Échap : passer (ne rien forcer,
  rester ici, pas de combat).

## Choix d'adaptation

Les points ambigus ou illisibles de la règle, et ce qui a été décidé :

- **Plateau** : la photo de l'encart, redressée. Une porte ou une fenêtre
  coupe parfois une case (on s'y place, « sur la case où se trouve
  l'obstacle », pour la forcer) ou passe entre deux cases (on la force
  depuis l'une ou l'autre). L'obstacle forcé ne vaut que pour le tour en
  cours (« l'effraction doit être effectuée même si l'obstacle a déjà été
  franchi »).
- **Pièces** : la règle ne les nomme pas ; elles ont reçu des noms
  (le hall, le grand salon…) pour le journal.
- **Meubles** : toujours placés au hasard par l'ordinateur, y compris à
  partir de quatre joueurs (la règle laisse alors les joueurs les
  placer). Les fenêtres peuvent être masquées par un meuble, pas les
  portes. Les armoires et les lits tiennent deux cases ; la tête du dormeur
  est sur la seconde. Les pions retournés restent visibles ; seul leur
  découvreur peut forcer un meuble, et s'il s'éloigne le meuble est vide.
- **Équipement** : le passe-partout compte comme un ou deux tours de plus
  sur la table d'effraction pour toute serrure (grille, portes, meubles),
  pas pour les fenêtres. Le revolver sert aussi de matraque, sans bruit.
- **Entrer** : on se poste sur une case-mur ; au tour suivant, on saute dans
  une case voisine du jardin (jet de saut) et le déplacement s'arrête là.
  Devant la grille, on peut la forcer dès le premier tour puis entrer à
  pied. On quitte la propriété depuis n'importe quelle case-mur.
- **Fenêtres** : fenêtre près d'un arbre (case voisine sous un arbre) : case
  normale ; sinon entrer coûte 2 points de déplacement, et sortir est un
  saut (table des combats) qui termine le déplacement.
- **Bruit** : distance en cases, à vol d'oiseau (les murs n'arrêtent pas le
  bruit). Le bruit de la marche est mesuré depuis la case d'arrivée et vaut
  le nombre de cases parcourues dans le tour.
- **Gardes** : ils ne montent pas sur le mur et ne passent jamais par les
  fenêtres ; en ronde ils restent dans la villa (la porte extérieure les
  arrête, comme le dit la règle), en alerte ou en poursuite ils passent les
  portes extérieures et vont dans le jardin. « Chaque joueur déplace les
  deux gardes d'une case » : une case par garde après le tour de chaque
  joueur (deux en alerte). Les chiffres autour des bases ont été relevés
  sur le plateau. Un garde voit un voleur quand ils sont dans la même pièce
  (ou tous deux dehors), et ne peut pas entrer sur la case d'un meuble.
- **Combats** : on attaque après s'être déplacé. Contre un garde, le garde
  riposte au revolver si le voleur le manque ; entre voleurs, on frappe à
  tour de rôle jusqu'au résultat. Assommé (KO) : deux tours sur place pour
  un voleur ; un garde ou un dormeur assommé reprend ses esprits au bout de
  deux tours (le dormeur donne alors l'alarme). Le combat n'est proposé au
  joueur humain que si un adversaire est à 3 cases ou moins.
- **Vu** : un voleur vu par un garde ou un dormeur réveillé, même sans avoir
  frappé personne, voit ses objets divisés par cinq (la règle ne le dit que
  pour « assommer », mais son barème parle de voleur « vu »).
- **Fin** : la partie s'arrête quand tous les voleurs sont sortis, arrêtés
  ou morts ; au lever du jour (tour 150) ceux qui restent sont pris.
- **Ordinateur** : il vise les meubles encore cachés, force ceux qu'il a
  découverts en réglant la durée pour que le bruit n'atteigne ni garde ni
  dormeur connu, ne force un obstacle que s'il rapproche de son but, et
  repart par le mur dès qu'il a assez de butin (son ambition varie d'un
  voleur à l'autre), que les gardes approchent après l'alarme, ou qu'il n'y
  a plus rien à chercher. Sur 60 parties : environ 60 % des voleurs
  ressortent (66 000 F en moyenne), les autres sont pris.

## Crédits et licence

*Fric-Frac* est un jeu de François Nédélec, paru en encart dans
*Jeux & Stratégie*. Le jeu original, sa règle et ses illustrations
appartiennent à leurs auteurs. Cette adaptation est un hommage non
commercial.

Les scans du magazine ne sont pas distribués dans ce dépôt. Seules les
images dont le jeu a besoin le sont (`assets/gen`).

Le code est sous licence MIT (voir `LICENSE`). Les polices de
`assets/polices` gardent leurs propres licences (fichiers joints) : SIL OFL
pour Limelight, Playfair Display et Josefin Sans, Apache 2.0 pour Special
Elite.
