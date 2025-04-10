package core

import "github.com/szcvak/sps/pkg/config"

func EmbedMilestones(stream *ByteStream) {
	goalIdx0 := config.MaximumRank - 1
	goalIdx5 := 499

	stream.Write(VInt(goalIdx0 + goalIdx5))

	for i := 0; i < goalIdx0; i++ {
		stream.Write(VInt(1))
		stream.Write(VInt(i))

		if i >= 34 {
			stream.Write(VInt(TrophyProgressStart[33]+50*(i-33)))
		} else {
			stream.Write(VInt(TrophyProgressStart[i]))
		}

		if i >= 34 {
			stream.Write(VInt(50))
		} else {
			stream.Write(VInt(TrophyProgress[i]))
		}

		stream.Write(VInt(0))
		stream.Write(VInt(1))
		stream.Write(VInt(1))
		stream.Write(VInt(10))
		stream.Write(VInt(5))
		stream.Write(VInt(1))
	}

	for i := 0; i < goalIdx5; i++ {
		stream.Write(VInt(5))
		stream.Write(VInt(i))

		stream.Write(VInt(ExperienceProgressStart[i]))
		stream.Write(VInt(ExperienceProgress[i]))

		stream.Write(VInt(0))
		stream.Write(VInt(1))
		stream.Write(VInt(12))
		stream.Write(VInt(20))

		stream.Write(ScId{5, 1})
	}
}

// --- Ugly code ahead --- //

var (
	TrophyProgressStart = []int{0,10,20,30,40,60,80,100,120,140,160,180,220,260,300,340,380,420,460,500,550,600,650,700,750,800,850,900,950,1000,1050,1100,1150,1200}
	TrophyProgress = []int{10,10,10,10,20,20,20,20,20,20,20,40,40,40,40,40,40,40,40,50,50,50,50,50,50,50,50,50,50,50,50,50,50,50,50}

	ExperienceProgressStart = []int{0,40,90,150,220,300,390,490,600,720,850,990,1140,1300,1470,1650,1840,2040,2250,2470,2700,2940,3190,3450,3720,4000,4290,4590,4900,5220,5550,5890,6240,6600,6970,7350,7740,8140,8550,8970,9400,9840,10290,10750,11220,11700,12190,12690,13200,13720,14250,14790,15340,15900,16470,17050,17640,18240,18850,19470,20100,20740,21390,22050,22720,23400,24090,24790,25500,26220,26950,27690,28440,29200,29970,30750,31540,32340,33150,33970,34800,35640,36490,37350,38220,39100,39990,40890,41800,42720,43650,44590,45540,46500,47470,48450,49440,50440,51450,52470,53500,54540,55590,56650,57720,58800,59890,60990,62100,63220,64350,65490,66640,67800,68970,70150,71340,72540,73750,74970,76200,77440,78690,79950,81220,82500,83790,85090,86400,87720,89050,90390,91740,93100,94470,95850,97240,98640,100050,101470,102900,104340,105790,107250,108720,110200,111690,113190,114700,116220,117750,119290,120840,122400,123970,125550,127140,128740,130350,131970,133600,135240,136890,138550,140220,141900,143590,145290,147000,148720,150450,152190,153940,155700,157470,159250,161040,162840,164650,166470,168300,170140,171990,173850,175720,177600,179490,181390,183300,185220,187150,189090,191040,193000,194970,196950,198940,200940,202950,204970,207000,209040,211090,213150,215220,217300,219390,221490,223600,225720,227850,229990,232140,234300,236470,238650,240840,243040,245250,247470,249700,251940,254190,256450,258720,261000,263290,265590,267900,270220,272550,274890,277240,279600,281970,284350,286740,289140,291550,293970,296400,298840,301290,303750,306220,308700,311190,313690,316200,318720,321250,323790,326340,328900,331470,334050,336640,339240,341850,344470,347100,349740,352390,355050,357720,360400,363090,365790,368500,371220,373950,376690,379440,382200,384970,387750,390540,393340,396150,398970,401800,404640,407490,410350,413220,416100,418990,421890,424800,427720,430650,433590,436540,439500,442470,445450,448440,451440,454450,457470,460500,463540,466590,469650,472720,475800,478890,481990,485100,488220,491350,494490,497640,500800,503970,507150,510340,513540,516750,519970,523200,526440,529690,532950,536220,539500,542790,546090,549400,552720,556050,559390,562740,566100,569470,572850,576240,579640,583050,586470,589900,593340,596790,600250,603720,607200,610690,614190,617700,621220,624750,628290,631840,635400,638970,642550,646140,649740,653350,656970,660600,664240,667890,671550,675220,678900,682590,686290,690000,693720,697450,701190,704940,708700,712470,716250,720040,723840,727650,731470,735300,739140,742990,746850,750720,754600,758490,762390,766300,770220,774150,778090,782040,786000,789970,793950,797940,801940,805950,809970,814000,818040,822090,826150,830220,834300,838390,842490,846600,850720,854850,858990,863140,867300,871470,875650,879840,884040,888250,892470,896700,900940,905190,909450,913720,918000,922290,926590,930900,935220,939550,943890,948240,952600,956970,961350,965740,970140,974550,978970,983400,987840,992290,996750,1001220,1005700,1010190,1014690,1019200,1023720,1028250,1032790,1037340,1041900,1046470,1051050,1055640,1060240,1064850,1069470,1074100,1078740,1083390,1088050,1092720,1097400,1102090,1106790,1111500,1116220,1120950,1125690,1130440,1135200,1139970,1144750,1149540,1154340,1159150,1163970,1168800,1173640,1178490,1183350,1188220,1193100,1197990,1202890,1207800,1212720,1217650,1222590,1227540,1232500,1237470,1242450,1247440,1252440,1257450}
	ExperienceProgress = []int{40,50,60,70,80,90,100,110,120,130,140,150,160,170,180,190,200,210,220,230,240,250,260,270,280,290,300,310,320,330,340,350,360,370,380,390,400,410,420,430,440,450,460,470,480,490,500,510,520,530,540,550,560,570,580,590,600,610,620,630,640,650,660,670,680,690,700,710,720,730,740,750,760,770,780,790,800,810,820,830,840,850,860,870,880,890,900,910,920,930,940,950,960,970,980,990,1000,1010,1020,1030,1040,1050,1060,1070,1080,1090,1100,1110,1120,1130,1140,1150,1160,1170,1180,1190,1200,1210,1220,1230,1240,1250,1260,1270,1280,1290,1300,1210,1320,1330,1340,1350,1360,1370,1380,1390,1400,1410,1420,1430,1440,1450,1460,1470,1480,1490,1500,1510,1520,1530,1540,1550,1560,1570,1580,1590,1600,1610,1620,1630,1640,1650,1660,1670,1680,1690,1700,1710,1720,1730,1740,1750,1760,1770,1780,1790,1800,1810,1820,1830,1840,1850,1860,1870,1880,1890,1900,1910,1920,1930,1940,1950,1960,1970,1980,1990,2000,2010,2020,2030,2040,2050,2060,2070,2080,2090,2100,2110,2120,2130,2140,2150,2160,2170,2180,2190,2200,2210,2220,2230,2240,2250,2260,2270,2280,2290,2300,2310,2320,2330,2340,2350,2360,2370,2380,2390,2400,2410,2420,2430,2440,2450,2460,2470,2480,2490,2500,2510,2520,2530,2540,2550,2560,2570,2580,2590,2600,2610,2620,2630,2640,2650,2660,2670,2680,2690,2700,2710,2720,2730,2740,2750,2760,2770,2780,2790,2800,2810,2820,2830,2840,2850,2860,2870,2880,2890,2900,2910,2920,2930,2940,2950,2960,2970,2980,2990,3000,3010,3020,3030,3040,3050,3060,3070,3080,3090,3100,3110,3120,3130,3140,3150,3160,3170,3180,3190,3200,3210,3220,3230,3240,3250,3260,3270,3280,3290,3300,3310,3320,3330,3340,3350,3360,3370,3380,3390,3400,3410,3420,3430,3440,3450,3460,3470,3480,3490,3500,3510,3520,3530,3540,3550,3560,3570,3580,3590,3600,3610,3620,3630,3640,3650,3660,3670,3680,3690,3700,3710,3720,3730,3740,3750,3760,3770,3780,3790,3800,3810,3820,3830,3840,3850,3860,3870,3880,3890,3900,3910,3920,3930,3940,3950,3960,3970,3980,3990,4000,4010,4020,4030,4040,4050,4060,4070,4080,4090,4100,4110,4120,4130,4140,4150,4160,4170,4180,4190,4200,4210,4220,4230,4240,4250,4260,4270,4280,4290,4300,4310,4320,4330,4340,4350,4360,4370,4380,4390,4400,4410,4420,4430,4440,4450,4460,4470,4480,4490,4500,4510,4520,4530,4540,4550,4560,4570,4580,4590,4600,4610,4620,4630,4640,4650,4660,4670,4680,4690,4700,4710,4720,4730,4740,4750,4760,4770,4780,4790,4800,4810,4820,4830,4840,4850,4860,4870,4880,4890,4900,4910,4920,4930,4940,4950,4960,4970,4980,4990,5000,5010,5020}
)
