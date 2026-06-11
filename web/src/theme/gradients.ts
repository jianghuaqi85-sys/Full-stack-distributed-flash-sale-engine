// Flat color palette for event cards — no gradients, no decoration
const eventColors = [
  '#0A0A0A', // near black
  '#1A1A2E', // dark navy
  '#16213E', // deep blue
  '#0F3460', // navy
  '#533483', // purple
  '#2C3E50', // slate
  '#34495E', // dark gray
  '#2C2C54', // dark violet
]

const cardIconColors = [
  '#0066FF', // blue
  '#10B981', // green
  '#F59E0B', // amber
  '#EF4444', // red
  '#8B5CF6', // violet
]

export function getEventColor(eventId: number): string {
  return eventColors[eventId % eventColors.length]
}

export function getCardIconColor(index: number): string {
  return cardIconColors[index % cardIconColors.length]
}

// Backward compatibility aliases
export const eventGradients = eventColors
export const cardIconGradients = cardIconColors
export const getEventGradient = getEventColor
export const getCardIconGradient = getCardIconColor
