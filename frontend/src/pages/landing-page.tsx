import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { ModeToggle } from '@/components/ModeToggle'
import { ThemeSelector } from '@/components/ThemeSelector'
import { 
  BookOpen, 
  Layers, 
  Lock, 
  Sparkles,
  ArrowRight,
  CheckCircle2,
  Users,
  FileText,
  FolderTree
} from 'lucide-react'

export function LandingPage() {
  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="fixed top-0 left-0 right-0 z-50 border-b border-border bg-background/80 backdrop-blur-md">
        <nav className="container mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <BookOpen className="h-6 w-6 text-primary" />
            <span className="font-semibold text-lg text-foreground">Notes</span>
          </div>
          <div className="flex items-center gap-3">
            <ThemeSelector />
            <ModeToggle />
            <Link to="/sign-in">
              <Button variant="ghost" size="sm">
                Sign In
              </Button>
            </Link>
            <Link to="/sign-up">
              <Button size="sm">
                Get Started
              </Button>
            </Link>
          </div>
        </nav>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-4 pt-32 pb-20">
        <div className="max-w-4xl mx-auto text-center">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-accent text-accent-foreground text-sm mb-8">
            <Sparkles className="h-4 w-4" />
            <span>Your thoughts, beautifully organized</span>
          </div>
          <h1 className="text-5xl md:text-6xl font-bold text-foreground mb-6 leading-tight">
            Write, organize, and
            <br />
            <span className="text-primary">share your ideas</span>
          </h1>
          <p className="text-xl text-muted-foreground mb-10 max-w-2xl mx-auto leading-relaxed">
            A workspace designed for clarity and focus. Structure your thoughts with notebooks, chapters, and notes that adapt to your workflow.
          </p>
          <div className="flex items-center justify-center gap-4">
            <Link to="/sign-up">
              <Button size="lg" className="gap-2">
                Start Writing
                <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
            <Link to="/sign-in">
              <Button size="lg" variant="outline">
                Sign In
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Features Grid */}
      <section className="container mx-auto px-4 py-20">
        <div className="max-w-6xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-foreground mb-4">
              Everything you need
            </h2>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
              Built for writers, thinkers, and teams who value organization and simplicity.
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
            {/* Feature 1 */}
            <div className="group p-6 rounded-lg border border-border bg-card hover:shadow-lg transition-all duration-300">
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <FolderTree className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-semibold text-card-foreground mb-2">
                Structured Organization
              </h3>
              <p className="text-muted-foreground">
                Create notebooks for projects, chapters for topics, and notes for individual ideas. Keep everything exactly where you need it.
              </p>
            </div>

            {/* Feature 2 */}
            <div className="group p-6 rounded-lg border border-border bg-card hover:shadow-lg transition-all duration-300">
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <FileText className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-semibold text-card-foreground mb-2">
                Powerful Writing
              </h3>
              <p className="text-muted-foreground">
                Focus on your ideas with a clean, distraction-free editor that gets out of your way and lets you write.
              </p>
            </div>

            {/* Feature 3 */}
            <div className="group p-6 rounded-lg border border-border bg-card hover:shadow-lg transition-all duration-300">
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <Layers className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-semibold text-card-foreground mb-2">
                Flexible Themes
              </h3>
              <p className="text-muted-foreground">
                Choose from multiple carefully crafted themes that match your style and make your workspace truly yours.
              </p>
            </div>

            {/* Feature 4 */}
            <div className="group p-6 rounded-lg border border-border bg-card hover:shadow-lg transition-all duration-300">
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <Users className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-semibold text-card-foreground mb-2">
                Share Your Work
              </h3>
              <p className="text-muted-foreground">
                Make your notebooks public and share your knowledge with others. Perfect for documentation, guides, and collaborative projects.
              </p>
            </div>

            {/* Feature 5 */}
            <div className="group p-6 rounded-lg border border-border bg-card hover:shadow-lg transition-all duration-300">
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <Lock className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-semibold text-card-foreground mb-2">
                Secure & Private
              </h3>
              <p className="text-muted-foreground">
                Your notes are protected with modern security practices. Control exactly what you share and what stays private.
              </p>
            </div>

            {/* Feature 6 */}
            <div className="group p-6 rounded-lg border border-border bg-card hover:shadow-lg transition-all duration-300">
              <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
                <Sparkles className="h-6 w-6 text-primary" />
              </div>
              <h3 className="text-xl font-semibold text-card-foreground mb-2">
                Smart Features
              </h3>
              <p className="text-muted-foreground">
                Powerful search, quick navigation, and keyboard shortcuts help you work faster and stay focused on what matters.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="container mx-auto px-4 py-20">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-4xl font-bold text-foreground mb-4">
              Simple by design
            </h2>
            <p className="text-lg text-muted-foreground">
              Get started in three easy steps
            </p>
          </div>

          <div className="space-y-12">
            {/* Step 1 */}
            <div className="flex gap-6 items-start">
              <div className="flex-shrink-0 h-12 w-12 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold text-lg">
                1
              </div>
              <div>
                <h3 className="text-xl font-semibold text-foreground mb-2">
                  Create your first notebook
                </h3>
                <p className="text-muted-foreground">
                  Start by creating a notebook for your project, class, or any topic you want to explore. Think of it as a container for related ideas.
                </p>
              </div>
            </div>

            {/* Step 2 */}
            <div className="flex gap-6 items-start">
              <div className="flex-shrink-0 h-12 w-12 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold text-lg">
                2
              </div>
              <div>
                <h3 className="text-xl font-semibold text-foreground mb-2">
                  Organize with chapters
                </h3>
                <p className="text-muted-foreground">
                  Break down your notebook into chapters. Each chapter can represent a subtopic, week, or any logical division that makes sense for you.
                </p>
              </div>
            </div>

            {/* Step 3 */}
            <div className="flex gap-6 items-start">
              <div className="flex-shrink-0 h-12 w-12 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold text-lg">
                3
              </div>
              <div>
                <h3 className="text-xl font-semibold text-foreground mb-2">
                  Write and refine
                </h3>
                <p className="text-muted-foreground">
                  Add notes to your chapters and start writing. Edit, reorganize, and refine your thoughts over time. Your ideas evolve with you.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Benefits Section */}
      <section className="container mx-auto px-4 py-20">
        <div className="max-w-5xl mx-auto">
          <div className="bg-gradient-to-br from-primary/10 to-primary/5 rounded-2xl p-12 border border-primary/20">
            <div className="max-w-3xl">
              <h2 className="text-3xl md:text-4xl font-bold text-foreground mb-6">
                Built for the way you think
              </h2>
              <p className="text-lg text-muted-foreground mb-8">
                Whether you're a student organizing class notes, a writer drafting your next piece, or a professional managing project documentation, our platform adapts to your needs.
              </p>
              <div className="grid md:grid-cols-2 gap-4 mb-8">
                <div className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
                  <span className="text-foreground">No learning curve required</span>
                </div>
                <div className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
                  <span className="text-foreground">Works on any device</span>
                </div>
                <div className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
                  <span className="text-foreground">Seamless organization</span>
                </div>
                <div className="flex items-start gap-3">
                  <CheckCircle2 className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
                  <span className="text-foreground">Always improving</span>
                </div>
              </div>
              <Link to="/sign-up">
                <Button size="lg" className="gap-2">
                  Get Started Now
                  <ArrowRight className="h-4 w-4" />
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="container mx-auto px-4 py-20">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-4xl md:text-5xl font-bold text-foreground mb-6">
            Start organizing your ideas today
          </h2>
          <p className="text-xl text-muted-foreground mb-10 max-w-2xl mx-auto">
            Join others who have found a better way to capture, organize, and share their thoughts.
          </p>
          <Link to="/sign-up">
            <Button size="lg" className="gap-2">
              Create Your Free Account
              <ArrowRight className="h-4 w-4" />
            </Button>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border bg-muted/50">
        <div className="container mx-auto px-4 py-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <BookOpen className="h-5 w-5 text-primary" />
              <span className="font-semibold text-foreground">Notes</span>
            </div>
            <div className="flex items-center gap-6 text-sm text-muted-foreground">
              <Link to="/sign-in" className="hover:text-foreground transition-colors">
                Sign In
              </Link>
              <Link to="/sign-up" className="hover:text-foreground transition-colors">
                Sign Up
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

