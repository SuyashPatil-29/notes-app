export {}

declare global {
  interface CustomJwtSessionClaims {
    metadata: {
      onboardingCompleted?: boolean
      onboardingType?: string
    }
  }
}

