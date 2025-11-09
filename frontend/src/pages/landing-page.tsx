import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { ModeToggle } from '@/components/ModeToggle'
import { ThemeSelector } from '@/components/ThemeSelector'
import { useTheme } from '@/components/theme-provider'
import {
  BookOpen,
  Sparkles,
  ArrowRight,
  Zap,
  Share2,
  Monitor,
  Smartphone,
  Tablet,
  Github,
  Twitter,
  Linkedin,
  Mail,
  Send
} from 'lucide-react'

// Theme to image mapping
const themeImages = {
  notebook: '/images/LandingImages/NotebookLandingImage.jpeg',
  tech: '/images/LandingImages/DarkMatterLandingImage.jpeg',
  minimal: '/images/LandingImages/GraphiteLandingImage.jpeg',
  orange: '/images/LandingImages/ClaudeLandingImage.jpeg',
  pink: '/images/LandingImages/T3ChatLandingImage.jpeg',
  supabase: '/images/LandingImages/SupabaseLandingImage.jpeg',
  gruvbox: '/images/LandingImages/GruvboxLandingImage.jpeg',
} as const

export function LandingPage() {
  const { theme } = useTheme()
  return (
    <div className="min-h-screen bg-background overflow-hidden">
      {/* Animated background gradient */}
      <div className="fixed inset-0 -z-10 bg-linear-to-br from-primary/5 via-background to-primary/10 animate-gradient-xy" />
      <div className="fixed inset-0 -z-10 bg-[radial-gradient(ellipse_at_top,var(--tw-gradient-stops))] from-primary/20 via-background to-background opacity-50" />

      {/* Floating orbs */}
      <div className="fixed top-20 left-10 w-72 h-72 bg-primary/20 rounded-full blur-3xl animate-float" />
      <div className="fixed bottom-20 right-10 w-96 h-96 bg-primary/10 rounded-full blur-3xl animate-float-delayed" />

      {/* Header */}
      <header className="fixed top-0 left-0 right-0 z-50 border-b border-border/40 bg-background/60 backdrop-blur-xl">
        <nav className="container mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2 group cursor-pointer">
            <BookOpen className="h-6 w-6 text-primary" />
            <Link to="/">
              <span className="font-bold text-lg bg-linear-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">Atlas</span>
            </Link>
          </div>
          <div className="flex items-center gap-3">
            <ThemeSelector />
            <ModeToggle />
            <Link to="/sign-in">
              <Button variant="ghost" size="sm" className="hover:scale-105 transition-transform">
                Sign In
              </Button>
            </Link>
            <Link to="/sign-up">
              <Button size="sm" className="hover:scale-105 transition-transform shadow-lg shadow-primary/25">
                Get Started
              </Button>
            </Link>
          </div>
        </nav>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-4 pt-32 pb-20 relative">
        <div className="max-w-5xl mx-auto text-center">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/10 border border-primary/20 text-sm mb-8 animate-fade-in-up backdrop-blur-sm">
            <Sparkles className="h-4 w-4 text-primary animate-pulse" />
            <span className="font-medium bg-linear-to-r from-primary to-primary/70 bg-clip-text text-transparent">
              AI-powered knowledge management
            </span>
          </div>

          <h1 className="text-5xl md:text-6xl font-bold text-foreground mb-6 leading-tight animate-fade-in-up">
            Your digital brain for
            <br />
            <span className="text-primary">everything that matters</span>
          </h1>

          <p className="text-xl text-muted-foreground mb-10 max-w-3xl mx-auto leading-relaxed animate-fade-in-up">
            Capture ideas instantly via WhatsApp, let AI organize them intelligently, and access your complete knowledge base from anywhere.
            From technical documentation to creative writing—Atlas transforms how you manage information.
          </p>

          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 animate-fade-in-up">
            <Link to="/sign-up">
              <Button size="lg" className="gap-2 text-lg px-8 py-6 hover:scale-105 transition-all shadow-2xl shadow-primary/25 hover:shadow-primary/40 group">
                Start Free Today
                <ArrowRight className="h-5 w-5 group-hover:translate-x-1 transition-transform" />
              </Button>
            </Link>
            <Button size="lg" variant="outline" className="gap-2 text-lg px-8 py-6 hover:scale-105 transition-all group backdrop-blur-sm">
              See Demo
            </Button>
          </div>
        </div>
      </section>

      {/* Audience Trust Bar */}
      <section className="container mx-auto px-4 pb-12">
        <div className="max-w-4xl mx-auto">
          <p className="text-center text-sm text-muted-foreground mb-4">Trusted by teams and individuals from</p>
          <div className="flex flex-wrap justify-center gap-3">
            {['Software Teams', 'Content Creators', 'Project Managers', 'Researchers', 'Students', 'Remote Teams'].map(audience => (
              <span key={audience} className="px-4 py-2 rounded-full bg-muted/50 hover:bg-muted transition-colors text-muted-foreground text-sm">
                {audience}
              </span>
            ))}
          </div>
        </div>
      </section>

      {/* Product Demo Section */}
      <section className="container mx-auto px-4 py-20">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-16 animate-fade-in-up">
            <h2 className="text-4xl md:text-5xl font-bold text-foreground mb-4">
              See it in action
            </h2>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              Experience the most intuitive note-taking workspace ever created
            </p>
          </div>

          {/* Main Product Screenshot */}
          <div className="relative animate-fade-in-up">
            <div className="absolute inset-0 bg-linear-to-r from-primary/20 to-primary/10 blur-3xl -z-10" />
            <div className="relative rounded-2xl overflow-hidden border border-border/50 shadow-2xl bg-card/50 backdrop-blur-sm">
              {/* Browser Chrome */}
              <div className="bg-muted/80 border-b border-border/50 px-4 py-3 flex items-center gap-2">
                <div className="flex gap-2">
                  <div className="w-3 h-3 rounded-full bg-red-500/80" />
                  <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
                  <div className="w-3 h-3 rounded-full bg-primary/80" />
                </div>
                <div className="flex-1 ml-4 bg-background/50 rounded px-3 py-1 text-xs text-muted-foreground">
                  tryatlas.ink/dashboard
                </div>
              </div>

              <img
                src={themeImages[theme as keyof typeof themeImages] || themeImages.notebook}
                alt={`Atlas Dashboard - ${theme} theme`}
                className="w-full h-auto object-cover rounded-lg"
              />
            </div>
          </div>

          {/* Multi-device showcase */}
          <div className="grid md:grid-cols-3 gap-6 mt-12 animate-fade-in-up">
            <div className="text-center p-6 rounded-xl bg-card/50 backdrop-blur-sm border border-border/50 hover:scale-105 transition-all group">
              <Monitor className="h-10 w-10 text-primary mx-auto mb-3 group-hover:scale-110 transition-transform" />
              <h4 className="font-semibold text-foreground mb-1">Desktop</h4>
              <p className="text-sm text-muted-foreground">Full-featured experience</p>
            </div>
            <div className="text-center p-6 rounded-xl bg-card/50 backdrop-blur-sm border border-border/50 hover:scale-105 transition-all group">
              <Tablet className="h-10 w-10 text-primary mx-auto mb-3 group-hover:scale-110 transition-transform" />
              <h4 className="font-semibold text-foreground mb-1">Tablet</h4>
              <p className="text-sm text-muted-foreground">Optimized for touch</p>
            </div>
            <div className="text-center p-6 rounded-xl bg-card/50 backdrop-blur-sm border border-border/50 hover:scale-105 transition-all group">
              <Smartphone className="h-10 w-10 text-primary mx-auto mb-3 group-hover:scale-110 transition-transform" />
              <h4 className="font-semibold text-foreground mb-1">Mobile</h4>
              <p className="text-sm text-muted-foreground">Notes on the go</p>
            </div>
          </div>
        </div>
      </section>

      {/* Features Bento Grid */}
      <section className="container mx-auto px-4 py-20">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-16 animate-fade-in-up">
            <h2 className="text-4xl md:text-5xl font-bold text-foreground mb-4">
              Built for how you actually work
            </h2>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              Real features that solve real problems—no fluff, no gimmicks
            </p>
          </div>

          {/* Bento Grid Layout */}
          <div className="grid md:grid-cols-3 gap-6 animate-fade-in-up">
            {/* Large Feature 1 - WhatsApp Integration */}
            <div className="md:col-span-2 md:row-span-2 relative group p-8 rounded-2xl border border-border bg-linear-to-br from-card via-card to-primary/5 hover:shadow-2xl hover:shadow-primary/10 transition-all duration-500 overflow-hidden">
              <div className="absolute top-0 right-0 w-64 h-64 bg-primary/10 rounded-full blur-3xl group-hover:scale-150 transition-transform duration-700" />
              <div className="relative z-10 h-full flex flex-col">
                <div className="inline-flex h-14 w-14 rounded-2xl bg-primary/10 items-center justify-center mb-4 group-hover:scale-110 group-hover:rotate-3 transition-all">
                  <Smartphone className="h-7 w-7 text-primary" />
                </div>
                <h3 className="text-3xl font-bold text-card-foreground mb-3">
                  Text a note, it's instantly saved
                </h3>
                <p className="text-lg text-muted-foreground mb-6">
                  Send notes via WhatsApp and they show up in your workspace automatically. No app switching, no friction. Actually useful when you're out and an idea hits.
                </p>

                {/* WhatsApp Example */}
                <div className="mt-auto space-y-3 bg-background/50 backdrop-blur-sm rounded-xl p-4 border border-border/50">
                  <div className="flex items-start gap-2">
                    <div className="w-6 h-6 rounded-full bg-primary/20 flex items-center justify-center shrink-0 mt-0.5">
                      <span className="text-xs">📱</span>
                    </div>
                    <div className="bg-primary/10 rounded-lg px-3 py-2 text-xs border border-primary/20">
                      Remember to refactor the auth flow tomorrow
                    </div>
                  </div>
                  <div className="text-xs text-muted-foreground pl-8">
                    → Saved to "Work Notes" notebook
                  </div>
                </div>
              </div>
            </div>

            {/* Feature 2 - Chat with your notes */}
            <div className="relative group p-6 rounded-2xl border border-border bg-card hover:shadow-xl transition-all duration-500 hover:-translate-y-1 flex flex-col">
              <div className="inline-flex h-12 w-12 rounded-xl bg-primary/10 items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <Sparkles className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-bold text-card-foreground mb-2">
                Ask questions across all your notes
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                "What did I write about the API redesign?" Get answers from your entire knowledge base in seconds.
              </p>
            </div>

            {/* Feature 3 - Linked notes */}
            <div className="relative group p-6 rounded-2xl border border-border bg-card hover:shadow-xl transition-all duration-500 hover:-translate-y-1">
              <div className="inline-flex h-12 w-12 rounded-xl bg-primary/10 items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <Share2 className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-bold text-card-foreground mb-2">
                Link notes together
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                Build a web of connected ideas. Type @ to link any note—your own personal wiki.
              </p>
            </div>

            {/* Feature 4 - Meeting notes */}
            <div className="md:col-span-3 relative group p-6 rounded-2xl border border-border bg-card hover:shadow-xl transition-all duration-500 hover:-translate-y-1">
              <div className="flex flex-col md:flex-row items-start gap-6">
                <div className="inline-flex h-12 w-12 rounded-xl bg-primary/10 items-center justify-center shrink-0 group-hover:scale-110 transition-transform">
                  <Zap className="h-6 w-6 text-primary" />
                </div>
                <div className="flex-1">
                  <h3 className="text-2xl font-bold text-card-foreground mb-3">
                    Auto-generated meeting notes
                  </h3>
                  <p className="text-muted-foreground mb-4">
                    Connect your calendar, and Atlas generates structured notes for your meetings. Captures action items, decisions, and key points so you can focus on the conversation instead of typing.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Use Cases Section */}
      <section className="container mx-auto px-4 py-20">
        <div className="max-w-5xl mx-auto">
          <div className="relative rounded-3xl overflow-hidden bg-linear-to-br from-primary/10 via-primary/5 to-background p-12 border border-primary/20 animate-fade-in-up">
            <div className="absolute inset-0 bg-grid-pattern opacity-5" />
            <div className="relative z-10">
              <div className="text-center mb-12">
                <h2 className="text-3xl md:text-4xl font-bold text-foreground mb-4">
                  Who actually uses this?
                </h2>
                <p className="text-lg text-muted-foreground">
                  Real people solving real problems with Atlas
                </p>
              </div>

              <div className="grid md:grid-cols-2 gap-6">
                <div className="p-6 rounded-xl bg-background/50 backdrop-blur-sm border border-border/50 hover:border-primary/30 transition-colors">
                  <div className="text-2xl mb-3">👨‍💻</div>
                  <h3 className="font-bold text-foreground mb-2">Developers</h3>
                  <p className="text-sm text-muted-foreground">
                    Keep track of bugs, code snippets, and architecture decisions. Link related notes together to build your own project wiki.
                  </p>
                </div>

                <div className="p-6 rounded-xl bg-background/50 backdrop-blur-sm border border-border/50 hover:border-primary/30 transition-colors">
                  <div className="text-2xl mb-3">📝</div>
                  <h3 className="font-bold text-foreground mb-2">Writers & Creators</h3>
                  <p className="text-sm text-muted-foreground">
                    Capture ideas the moment they hit. Organize research, drafts, and references in one place without worrying about file systems.
                  </p>
                </div>

                <div className="p-6 rounded-xl bg-background/50 backdrop-blur-sm border border-border/50 hover:border-primary/30 transition-colors">
                  <div className="text-2xl mb-3">🎯</div>
                  <h3 className="font-bold text-foreground mb-2">Product Managers</h3>
                  <p className="text-sm text-muted-foreground">
                    Meeting notes that auto-generate from your calendar. Ask AI to summarize decisions across multiple meetings.
                  </p>
                </div>

                <div className="p-6 rounded-xl bg-background/50 backdrop-blur-sm border border-border/50 hover:border-primary/30 transition-colors">
                  <div className="text-2xl mb-3">🎓</div>
                  <h3 className="font-bold text-foreground mb-2">Students & Researchers</h3>
                  <p className="text-sm text-muted-foreground">
                    Build a personal knowledge base that grows with you. Link concepts, organize by subject, and search across everything you've learned.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Final CTA Section */}
      <section className="container mx-auto px-4 py-32">
        <div className="max-w-4xl mx-auto text-center relative">
          {/* Glow effect */}
          <div className="absolute inset-0 bg-linear-to-r from-primary/20 via-primary/30 to-primary/20 blur-3xl -z-10 animate-pulse-slow" />

          <div className="animate-fade-in-up">
            <h2 className="text-4xl md:text-5xl lg:text-6xl font-black text-foreground mb-6 leading-tight">
              Ready to get started?
            </h2>
            <p className="text-xl md:text-2xl text-muted-foreground mb-12 max-w-2xl mx-auto">
              Join thousands who've already found a better way to organize their thoughts.
            </p>

            <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-8">
              <Link to="/sign-up">
                <Button size="lg" className="gap-2 text-lg px-10 py-7 hover:scale-105 transition-all shadow-2xl shadow-primary/25 hover:shadow-primary/40 group">
                  Create Your Free Account
                  <ArrowRight className="h-5 w-5 group-hover:translate-x-1 transition-transform" />
                </Button>
              </Link>
            </div>

            <p className="text-sm text-muted-foreground">
              No credit card required • Free forever • Set up in 30 seconds
            </p>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border/50 bg-card/30 backdrop-blur-sm">
        <div className="container mx-auto px-4 py-16">
          <div className="grid md:grid-cols-2 lg:grid-cols-5 gap-12 mb-12">
            {/* Brand Column */}
            <div className="lg:col-span-2">
              <div className="flex items-center gap-2 mb-4 group">
                <BookOpen className="h-7 w-7 text-primary transition-transform group-hover:scale-110 group-hover:rotate-12" />
                <span className="font-bold text-2xl bg-linear-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
                  Atlas
                </span>
              </div>
              <p className="text-muted-foreground mb-6 max-w-sm">
                The ultimate workspace for organizing your thoughts. Write, collaborate, and share your ideas with the world.
              </p>

              {/* Newsletter Signup */}
              <div className="space-y-3">
                <p className="text-sm font-semibold text-foreground">Stay Updated</p>
                <div className="flex gap-2">
                  <div className="relative flex-1">
                    <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                    <input
                      type="email"
                      placeholder="Enter your email"
                      className="w-full pl-10 pr-4 py-2 bg-background border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
                    />
                  </div>
                  <Button size="sm" className="px-4">
                    <Send className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>

            {/* Product Column */}
            <div>
              <h3 className="font-semibold text-foreground mb-4">Product</h3>
              <ul className="space-y-3">
                <li>
                  <a href="#features" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Features
                  </a>
                </li>
                <li>
                  <a href="#pricing" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Pricing
                  </a>
                </li>
                <li>
                  <a href="#security" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Security
                  </a>
                </li>
                <li>
                  <a href="#roadmap" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Roadmap
                  </a>
                </li>
                <li>
                  <a href="#changelog" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Changelog
                  </a>
                </li>
              </ul>
            </div>

            {/* Resources Column */}
            <div>
              <h3 className="font-semibold text-foreground mb-4">Resources</h3>
              <ul className="space-y-3">
                <li>
                  <a href="#docs" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Documentation
                  </a>
                </li>
                <li>
                  <a href="#guides" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Guides
                  </a>
                </li>
                <li>
                  <a href="#api" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    API Reference
                  </a>
                </li>
                <li>
                  <a href="#blog" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Blog
                  </a>
                </li>
                <li>
                  <a href="#support" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Support
                  </a>
                </li>
              </ul>
            </div>

            {/* Company Column */}
            <div>
              <h3 className="font-semibold text-foreground mb-4">Company</h3>
              <ul className="space-y-3">
                <li>
                  <a href="#about" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    About Us
                  </a>
                </li>
                <li>
                  <a href="#careers" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Careers
                  </a>
                </li>
                <li>
                  <Link to="/privacy" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Privacy Policy
                  </Link>
                </li>
                <li>
                  <Link to="/terms" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Terms of Service
                  </Link>
                </li>
                <li>
                  <a href="#contact" className="text-sm text-muted-foreground hover:text-primary transition-colors">
                    Contact
                  </a>
                </li>
              </ul>
            </div>
          </div>

          {/* Bottom Bar */}
          <div className="pt-8 border-t border-border/50 flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="text-sm text-muted-foreground">
              © 2025 Atlas. Built with ❤️ for writers and thinkers everywhere.
            </p>

            {/* Social Links */}
            <div className="flex items-center gap-4">
              <a
                href="https://github.com"
                target="_blank"
                rel="noopener noreferrer"
                className="w-9 h-9 rounded-lg bg-muted/50 hover:bg-primary/10 flex items-center justify-center transition-all hover:scale-110 group"
              >
                <Github className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
              </a>
              <a
                href="https://twitter.com"
                target="_blank"
                rel="noopener noreferrer"
                className="w-9 h-9 rounded-lg bg-muted/50 hover:bg-primary/10 flex items-center justify-center transition-all hover:scale-110 group"
              >
                <Twitter className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
              </a>
              <a
                href="https://linkedin.com"
                target="_blank"
                rel="noopener noreferrer"
                className="w-9 h-9 rounded-lg bg-muted/50 hover:bg-primary/10 flex items-center justify-center transition-all hover:scale-110 group"
              >
                <Linkedin className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

