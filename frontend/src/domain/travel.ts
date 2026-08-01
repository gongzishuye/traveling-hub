export type FoodId = 'light-meal' | 'picnic-basket'
export type JourneyPhase = 'home' | 'travelling' | 'returned'

export type Postcard = {
  id:
    | 'willow-pond'
    | 'mist-tea-slope'
    | 'firefly-ravine'
    | 'cloud-ridge'
    | 'old-bridge-market'
    | 'starfall-camp'
  title: string
  body: string
  alt: string
}
