export const eventGradients = [
  'linear-gradient(135deg, #5B2FE8 0%, #8B6FFF 100%)',
  'linear-gradient(135deg, #D4A843 0%, #F5C862 100%)',
  'linear-gradient(135deg, #6366F1 0%, #818CF8 100%)',
  'linear-gradient(135deg, #EC4899 0%, #F472B6 100%)',
  'linear-gradient(135deg, #14B8A6 0%, #2DD4BF 100%)',
  'linear-gradient(135deg, #F97316 0%, #FB923C 100%)',
  'linear-gradient(135deg, #8B5CF6 0%, #A78BFA 100%)',
  'linear-gradient(135deg, #06B6D4 0%, #22D3EE 100%)',
]

export const cardIconGradients = [
  'linear-gradient(135deg, #5B2FE8, #8B6FFF)',
  'linear-gradient(135deg, #6366F1, #818CF8)',
  'linear-gradient(135deg, #D4A843, #F5C862)',
  'linear-gradient(135deg, #14B8A6, #2DD4BF)',
  'linear-gradient(135deg, #EC4899, #F472B6)',
]

export function getEventGradient(eventId: number): string {
  return eventGradients[eventId % eventGradients.length]
}

export function getCardIconGradient(index: number): string {
  return cardIconGradients[index % cardIconGradients.length]
}
