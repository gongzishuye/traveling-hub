import type { FoodId, JourneyPhase, Postcard } from '../domain/travel'
import foodLightMeal from '../assets/props/light-meal.png'
import foodPicnicBasket from '../assets/props/picnic-basket.png'
import postcardCloudRidge from '../assets/postcards/cloud-ridge.png'
import postcardFireflyRavine from '../assets/postcards/firefly-ravine.png'
import postcardMistTeaSlope from '../assets/postcards/mist-tea-slope.png'
import postcardOldBridgeMarket from '../assets/postcards/old-bridge-market.png'
import postcardStarfallCamp from '../assets/postcards/starfall-camp.png'
import postcardWillowPond from '../assets/postcards/willow-pond.png'
import sceneHome from '../assets/scenes/garden-travelling-consistent-v2.png'
import sceneReturned from '../assets/scenes/garden-returned-consistent-v2.png'
import sceneTravelling from '../assets/scenes/garden-travelling-consistent-v2.png'
import detailHerbGarden from '../assets/illustrations/detail-herb-garden.svg?no-inline'
import detailLanternPost from '../assets/illustrations/detail-lantern-post.svg?no-inline'
import detailMailbox from '../assets/illustrations/detail-mailbox.svg?no-inline'
import detailMossBridge from '../assets/illustrations/detail-moss-bridge.svg?no-inline'
import detailTrailSign from '../assets/illustrations/detail-trail-sign.svg?no-inline'
import detailWindBell from '../assets/illustrations/detail-wind-bell.svg?no-inline'

type PostcardId = Postcard['id']

export type ArtVisual = {
  src: string
  alt: string
  description: string
}

export type DetailId =
  | 'herb-garden'
  | 'lantern-post'
  | 'mailbox'
  | 'moss-bridge'
  | 'trail-sign'
  | 'wind-bell'

export type WorldDetail = ArtVisual & {
  title: string
}

export type SceneVisual = ArtVisual & {
  detailIds: readonly [DetailId, DetailId, DetailId]
}

export const foodVisuals = {
  'light-meal': {
    src: foodLightMeal,
    alt: '木制食盒里的蔬菜饭团与热茶插画',
    description: '清爽饭团和热茶，为短途散步准备的轻食。',
  },
  'picnic-basket': {
    src: foodPicnicBasket,
    alt: '装着面包浆果和布巾的野餐篮插画',
    description: '盛有面包与浆果的编织篮，适合较远的午后旅程。',
  },
} satisfies Record<FoodId, ArtVisual>

export const postcardVisuals = {
  'willow-pond': {
    src: postcardWillowPond,
    alt: '柳枝石桥和静水池塘的旅行卡插画',
    description: '柳影映在池水里，石桥把岸边与远处连在一起。',
  },
  'cloud-ridge': {
    src: postcardCloudRidge,
    alt: '云雾松林与层叠山脊的旅行卡插画',
    description: '云雾掠过松林和山脊，留下干燥松香的记忆。',
  },
  'mist-tea-slope': {
    src: postcardMistTeaSlope,
    alt: '晨雾笼罩茶坡与石上茶壶的旅行卡插画',
    description: '茶垄沿着山坡展开，石台上的小壶把晨雾暖成一缕白汽。',
  },
  'firefly-ravine': {
    src: postcardFireflyRavine,
    alt: '暮色溪谷中萤火与枝头提灯的旅行卡插画',
    description: '蕨叶、溪石和枝头提灯陪着萤火，把溪谷照得温柔而安静。',
  },
  'old-bridge-market': {
    src: postcardOldBridgeMarket,
    alt: '旧木桥集市的香料摊与彩带旅行卡插画',
    description: '旧桥上的香料瓶、布带和打盹的猫，共享一段有阳光的午后。',
  },
  'starfall-camp': {
    src: postcardStarfallCamp,
    alt: '星空营地帐篷篝火与流星的旅行卡插画',
    description: '帐篷、篝火和树桩上的地图安静等待一颗越过松林的流星。',
  },
} satisfies Record<PostcardId, ArtVisual>

export const worldDetails = {
  'moss-bridge': {
    title: '苔藓石桥',
    src: detailMossBridge,
    alt: '覆盖嫩绿苔藓的小石桥细节插画',
    description: '溪水上短短的石桥，桥面长着柔软的嫩绿苔藓。',
  },
  'lantern-post': {
    title: '暖光路灯',
    src: detailLanternPost,
    alt: '木杆上亮起暖色玻璃路灯细节插画',
    description: '傍晚亮起的玻璃路灯，为回家的石路留一束暖光。',
  },
  'herb-garden': {
    title: '香草花圃',
    src: detailHerbGarden,
    alt: '木边框里的香草和小花细节插画',
    description: '小屋旁的香草花圃，生长着低矮叶片与浅黄小花。',
  },
  'trail-sign': {
    title: '林间路标',
    src: detailTrailSign,
    alt: '指向池塘和山谷的木制路标细节插画',
    description: '木制路标安静地指向池塘与通往山谷的远行小路。',
  },
  'wind-bell': {
    title: '风铃枝条',
    src: detailWindBell,
    alt: '树枝上悬着陶制风铃的细节插画',
    description: '陶制风铃挂在枝头，风经过时发出短促清亮的声音。',
  },
  mailbox: {
    title: '旅行信箱',
    src: detailMailbox,
    alt: '插着旅行卡的绿色木制信箱细节插画',
    description: '绿色木信箱露出一张旅行卡，收好归来后的远方消息。',
  },
} satisfies Record<DetailId, WorldDetail>

export const sceneVisuals = {
  home: {
    src: sceneHome,
    alt: '完整远景庭院中苔藓小屋前整理手账的原创小龙虾旅行家插画',
    description: '出发前的完整远景庭院铺开小屋、温室、花圃、石径与远山；原创小龙虾旅行家在手账桌旁整理下一段远行。',
    detailIds: ['moss-bridge', 'lantern-post', 'herb-garden'],
  },
  travelling: {
    src: sceneTravelling,
    alt: '旅人离开后无角色的完整远景庭院，空椅与打开手账留在小屋前的插画',
    description: '旅人已经出发，无角色的完整远景庭院只留下打开的手账、空椅和沿花圃通往远山的石径。',
    detailIds: ['moss-bridge', 'trail-sign', 'wind-bell'],
  },
  returned: {
    src: sceneReturned,
    alt: '暮色完整远景庭院中带着行囊归来的原创小龙虾旅行家插画',
    description: '完整远景庭院的暖灯映着小屋、温室与石径；原创小龙虾旅行家带着背包和旅行卡归来。',
    detailIds: ['lantern-post', 'herb-garden', 'mailbox'],
  },
} satisfies Record<JourneyPhase, SceneVisual>

export function preloadSceneHeroes() {
  for (const scene of Object.values(sceneVisuals)) {
    const image = new Image()
    image.src = scene.src
  }
}
