package journey

var templates = []Template{
	{ID: "willow-pond-reed", FoodID: "light-meal", PostcardID: "willow-pond", Events: [3]string{"芦苇在浅水边给旅人让出一条小路。", "一只青蜓停在石桥栏杆上，翅膀像薄玻璃。", "旅人从柳影池带回了一张带着水汽的旅行卡。"}},
	{ID: "willow-pond-boat", FoodID: "light-meal", PostcardID: "willow-pond", Events: [3]string{"岸边系着的小木船轻轻碰了碰木桩。", "倒映的云朵被桨尖划成几段缓慢的波纹。", "旅人从柳影池带回了一张写着舟声的旅行卡。"}},
	{ID: "willow-pond-rain", FoodID: "light-meal", PostcardID: "willow-pond", Events: [3]string{"细雨先落在荷叶上，随后才落进池里。", "石桥下的水面收集起一圈又一圈圆响。", "旅人从柳影池带回了一张沾着雨点的旅行卡。"}},
	{ID: "mist-tea-slope-sprout", FoodID: "light-meal", PostcardID: "mist-tea-slope", Events: [3]string{"雾从茶垄间升起，掩住了更高处的石阶。", "新芽在晨光里翻出一层柔软的浅绿。", "旅人从薄雾茶坡带回了一张留着新芽香气的旅行卡。"}},
	{ID: "mist-tea-slope-kettle", FoodID: "light-meal", PostcardID: "mist-tea-slope", Events: [3]string{"石台上的小壶冒出一缕安静的白汽。", "茶杯边缘映着山谷里缓慢移动的雾。", "旅人从薄雾茶坡带回了一张暖茶色的旅行卡。"}},
	{ID: "mist-tea-slope-bell", FoodID: "light-meal", PostcardID: "mist-tea-slope", Events: [3]string{"坡顶的风铃响了三下，又停回雾里。", "沿着木栈道走，脚步声被湿叶轻轻接住。", "旅人从薄雾茶坡带回了一张记着铃声的旅行卡。"}},
	{ID: "firefly-ravine-fern", FoodID: "light-meal", PostcardID: "firefly-ravine", Events: [3]string{"蕨叶从石缝里伸出来，指向溪谷的深处。", "溪水绕过圆石，把天色揉成细碎的蓝。", "旅人从萤灯溪谷带回了一张夹着蕨叶的旅行卡。"}},
	{ID: "firefly-ravine-lantern", FoodID: "light-meal", PostcardID: "firefly-ravine", Events: [3]string{"枝头的提灯先亮起，给小径留下一点金色。", "几只萤火虫绕着灯影，像在练习一段舞步。", "旅人从萤灯溪谷带回了一张照亮边角的旅行卡。"}},
	{ID: "firefly-ravine-song", FoodID: "light-meal", PostcardID: "firefly-ravine", Events: [3]string{"谷底传来一段短短的口哨，和溪声交错。", "晚风拨过树冠，萤火在暗处一盏盏醒来。", "旅人从萤灯溪谷带回了一张写着夜歌的旅行卡。"}},
	{ID: "cloud-ridge-pine", FoodID: "picnic-basket", PostcardID: "cloud-ridge", Events: [3]string{"松枝在高处相碰，落下一阵干净的松香。", "旅人把餐布铺在背风的平石上，望见远山。", "旅人从云脊坡带回了一张带着松针的旅行卡。"}},
	{ID: "cloud-ridge-cloud", FoodID: "picnic-basket", PostcardID: "cloud-ridge", Events: [3]string{"一团低云翻过山脊，短暂遮住了小径。", "云开后，坡下的河流像一根发亮的线。", "旅人从云脊坡带回了一张收着云影的旅行卡。"}},
	{ID: "cloud-ridge-stone", FoodID: "picnic-basket", PostcardID: "cloud-ridge", Events: [3]string{"路边的叠石上长着深色苔藓和一株小蕨。", "风停的片刻，远处山雀从岩边跳到了枝头。", "旅人从云脊坡带回了一张压着小石纹的旅行卡。"}},
	{ID: "old-bridge-market-spice", FoodID: "picnic-basket", PostcardID: "old-bridge-market", Events: [3]string{"旧桥尽头的香料摊飘来橙皮和八角的气味。", "摊主把一小包草籽系在野餐篮的提手上。", "旅人从旧桥集市带回了一张染着香料色的旅行卡。"}},
	{ID: "old-bridge-market-ribbon", FoodID: "picnic-basket", PostcardID: "old-bridge-market", Events: [3]string{"桥栏上的布带被风吹成一串柔软的小旗。", "午后的阳光穿过彩带，在木板上留下移动的影子。", "旅人从旧桥集市带回了一张系着布带的旅行卡。"}},
	{ID: "old-bridge-market-cat", FoodID: "picnic-basket", PostcardID: "old-bridge-market", Events: [3]string{"一只橘猫在装满果皮的木箱上晒着太阳。", "它醒来伸了个懒腰，又把尾巴盖回鼻尖。", "旅人从旧桥集市带回了一张印着猫爪的旅行卡。"}},
	{ID: "starfall-camp-tent", FoodID: "picnic-basket", PostcardID: "starfall-camp", Events: [3]string{"松林空地里，一顶小帐篷正被晚风吹得轻响。", "旅人把野餐篮放在火边，烤热了面包。", "旅人从星落营地带回了一张留着火光的旅行卡。"}},
	{ID: "starfall-camp-map", FoodID: "picnic-basket", PostcardID: "starfall-camp", Events: [3]string{"树桩上的旧地图摊开，边角压着一颗圆润的石子。", "地图上的溪流和营火旁的溪声刚好重合。", "旅人从星落营地带回了一张折着地图角的旅行卡。"}},
	{ID: "starfall-camp-comet", FoodID: "picnic-basket", PostcardID: "starfall-camp", Events: [3]string{"夜色落下时，第一颗星在帐篷上方亮起来。", "一颗长尾流星越过山脊，照亮了短短一瞬。", "旅人从星落营地带回了一张映着流星的旅行卡。"}},
}

func Templates() []Template {
	result := make([]Template, len(templates))
	copy(result, templates)
	return result
}

func TemplateByID(id string) (Template, bool) {
	for _, template := range templates {
		if template.ID == id {
			return template, true
		}
	}
	return Template{}, false
}
