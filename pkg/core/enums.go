package core

type PlayerState int32

const (
	StateSession PlayerState = iota
	StateLogin
	StateLoggedIn
)

var (
	RequiredExp = []int32{0, 40, 90, 150, 220, 300, 390, 490, 600, 720, 850, 990, 1140, 1300, 1470, 1650, 1840, 2040, 2250, 2470, 2700, 2940, 3190, 3450, 3720, 4000, 4290, 4590, 4900, 5220, 5550, 5890, 6240, 6600, 6970, 7350, 7740, 8140, 8550, 8970, 9400, 9840, 10290, 10750, 11220, 11700, 12190, 12690, 13200, 13720, 14250, 14790, 15340, 15900, 16470, 17050, 17640, 18240, 18850, 19470, 20100, 20740, 21390, 22050, 22720, 23400, 24090, 24790, 25500, 26220, 26950, 27690, 28440, 29200, 29970, 30750, 31540, 32340, 33150, 33970, 34800, 35640, 36490, 37350, 38220, 39100, 39990, 40890, 41800, 42720, 43650, 44590, 45540, 46500, 47470, 48450, 49440, 50440, 51450, 52470, 53500, 54540, 55590, 56650, 57720, 58800, 59890, 60990, 62100, 63220, 64350, 65490, 66640, 67800, 68970, 70150, 71340, 72540, 73750, 74970, 76200, 77440, 78690, 79950, 81220, 82500, 83790, 85090, 86400, 87720, 89050, 90390, 91740, 93100, 94470, 95850, 97240, 98640, 100050, 101470, 102900, 104340, 105790, 107250, 108720, 110200, 111690, 113190, 114700, 116220, 117750, 119290, 120840, 122400, 123970, 125550, 127140, 128740, 130350, 131970, 133600, 135240, 136890, 138550, 140220, 141900, 143590, 145290, 147000, 148720, 150450, 152190, 153940, 155700, 157470, 159250, 161040, 162840, 164650, 166470, 168300, 170140, 171990, 173850, 175720, 177600, 179490, 181390, 183300, 185220, 187150, 189090, 191040, 193000, 194970, 196950, 198940, 200940, 202950, 204970, 207000, 209040, 211090, 213150, 215220, 217300, 219390, 221490, 223600, 225720, 227850, 229990, 232140, 234300, 236470, 238650, 240840, 243040, 245250, 247470, 249700, 251940, 254190, 256450, 258720, 261000, 263290, 265590, 267900, 270220, 272550, 274890, 277240, 279600, 281970, 284350, 286740, 289140, 291550, 293970, 296400, 298840, 301290, 303750, 306220, 308700, 311190, 313690, 316200, 318720, 321250, 323790, 326340, 328900, 331470, 334050, 336640, 339240, 341850, 344470, 347100, 349740, 352390, 355050, 357720, 360400, 363090, 365790, 368500, 371220, 373950, 376690, 379440, 382200, 384970, 387750, 390540, 393340, 396150, 398970, 401800, 404640, 407490, 410350, 413220, 416100, 418990, 421890, 424800, 427720, 430650, 433590, 436540, 439500, 442470, 445450, 448440, 451440, 454450, 457470, 460500, 463540, 466590, 469650, 472720, 475800, 478890, 481990, 485100, 488220, 491350, 494490, 497640, 500800, 503970, 507150, 510340, 513540, 516750, 519970, 523200, 526440, 529690, 532950, 536220, 539500, 542790, 546090, 549400, 552720, 556050, 559390, 562740, 566100, 569470, 572850, 576240, 579640, 583050, 586470, 589900, 593340, 596790, 600250, 603720, 607200, 610690, 614190, 617700, 621220, 624750, 628290, 631840, 635400, 638970, 642550, 646140, 649740, 653350, 656970, 660600, 664240, 667890, 671550, 675220, 678900, 682590, 686290, 690000, 693720, 697450, 701190, 704940, 708700, 712470, 716250, 720040, 723840, 727650, 731470, 735300, 739140, 742990, 746850, 750720, 754600, 758490, 762390, 766300, 770220, 774150, 778090, 782040, 786000, 789970, 793950, 797940, 801940, 805950, 809970, 814000, 818040, 822090, 826150, 830220, 834300, 838390, 842490, 846600, 850720, 854850, 858990, 863140, 867300, 871470, 875650, 879840, 884040, 888250, 892470, 896700, 900940, 905190, 909450, 913720, 918000, 922290, 926590, 930900, 935220, 939550, 943890, 948240, 952600, 956970, 961350, 965740, 970140, 974550, 978970, 983400, 987840, 992290, 996750, 1001220, 1005700, 1010190, 1014690, 1019200, 1023720, 1028250, 1032790, 1037340, 1041900, 1046470, 1051050, 1055640, 1060240, 1064850, 1069470, 1074100, 1078740, 1083390, 1088050, 1092720, 1097400, 1102090, 1106790, 1111500, 1116220, 1120950, 1125690, 1130440, 1135200, 1139970, 1144750, 1149540, 1154340, 1159150, 1163970, 1168800, 1173640, 1178490, 1183350, 1188220, 1193100, 1197990, 1202890, 1207800, 1212720, 1217650, 1222590, 1227540, 1232500, 1237470, 1242450, 1247440, 1252440, 1257450}
)
