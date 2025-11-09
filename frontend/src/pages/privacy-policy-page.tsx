import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { BookOpen, ArrowLeft, Shield } from 'lucide-react'

export function PrivacyPolicyPage() {
  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b border-border/40 bg-background/80 backdrop-blur-xl">
        <nav className="container mx-auto px-4 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 group cursor-pointer">
            <BookOpen className="h-6 w-6 text-primary" />
            <span className="font-bold text-xl bg-linear-to-r from-primary to-primary/80 bg-clip-text text-transparent">
              Atlas
            </span>
          </Link>
          
          <Link to="/">
            <Button variant="ghost" className="gap-2">
              <ArrowLeft className="h-4 w-4" />
              Back to Home
            </Button>
          </Link>
        </nav>
      </header>

      {/* Content */}
      <main className="container mx-auto px-4 py-12 max-w-4xl">
        {/* Hero Section */}
        <div className="text-center mb-12">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary/10 mb-4">
            <Shield className="h-8 w-8 text-primary" />
          </div>
          <h1 className="text-4xl font-bold mb-4">Privacy Policy</h1>
          <p className="text-muted-foreground text-lg">
            Last updated: {new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
          </p>
        </div>

        {/* Policy Content */}
        <div className="prose prose-slate dark:prose-invert max-w-none">
          <div className="space-y-8">
            {/* Introduction */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Introduction</h2>
              <p className="text-muted-foreground leading-relaxed">
                Welcome to Atlas. We respect your privacy and are committed to protecting your personal data. 
                This privacy policy will inform you about how we look after your personal data when you visit 
                our website and use our services, and tell you about your privacy rights and how the law protects you.
              </p>
            </section>

            {/* Information We Collect */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Information We Collect</h2>
              <div className="space-y-4">
                <div>
                  <h3 className="text-xl font-semibold mb-2">Personal Information</h3>
                  <p className="text-muted-foreground leading-relaxed mb-2">
                    We collect personal information that you voluntarily provide to us when you:
                  </p>
                  <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                    <li>Register for an account</li>
                    <li>Create or edit notes, notebooks, and chapters</li>
                    <li>Use our AI-powered features</li>
                    <li>Communicate with us</li>
                    <li>Subscribe to our newsletter</li>
                  </ul>
                </div>

                <div>
                  <h3 className="text-xl font-semibold mb-2">Data You Create</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    This includes all content you create, upload, or store in Atlas, such as:
                  </p>
                  <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                    <li>Notes and their content</li>
                    <li>Notebooks and chapters</li>
                    <li>Task boards and tasks</li>
                    <li>Meeting transcripts and notes</li>
                    <li>Links and relationships between notes</li>
                  </ul>
                </div>

                <div>
                  <h3 className="text-xl font-semibold mb-2">Automatically Collected Information</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    When you use our services, we automatically collect certain information, including:
                  </p>
                  <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                    <li>Device information (browser type, operating system)</li>
                    <li>Usage data (features used, time spent)</li>
                    <li>IP address and location data</li>
                    <li>Cookies and similar tracking technologies</li>
                  </ul>
                </div>
              </div>
            </section>

            {/* How We Use Your Information */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">How We Use Your Information</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                We use your personal information for the following purposes:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                <li>To provide and maintain our services</li>
                <li>To improve and personalize your experience</li>
                <li>To process your requests and transactions</li>
                <li>To send you updates and notifications</li>
                <li>To provide AI-powered features and suggestions</li>
                <li>To detect and prevent fraud and abuse</li>
                <li>To comply with legal obligations</li>
                <li>To analyze usage patterns and improve our services</li>
              </ul>
            </section>

            {/* Data Storage and Security */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Data Storage and Security</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                We implement appropriate technical and organizational measures to protect your personal data:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                <li>End-to-end encryption for sensitive data</li>
                <li>Secure data storage with regular backups</li>
                <li>Access controls and authentication mechanisms</li>
                <li>Regular security audits and monitoring</li>
                <li>Compliance with industry-standard security practices</li>
              </ul>
              <p className="text-muted-foreground leading-relaxed mt-4">
                Your data is stored securely and is only accessible to authorized personnel who need access 
                to perform their job functions.
              </p>
            </section>

            {/* Third-Party Services */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Third-Party Services</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                We use third-party services to help us provide and improve our services:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                <li><strong>Authentication:</strong> Clerk for secure user authentication</li>
                <li><strong>AI Services:</strong> OpenAI, Anthropic, and Google AI for AI-powered features</li>
                <li><strong>Database:</strong> Supabase for data storage and real-time collaboration</li>
                <li><strong>Analytics:</strong> To understand how users interact with our service</li>
                <li><strong>Hosting:</strong> Cloud hosting providers for service delivery</li>
              </ul>
              <p className="text-muted-foreground leading-relaxed mt-4">
                These third parties have access to your data only to perform specific tasks on our behalf 
                and are obligated not to disclose or use it for any other purpose.
              </p>
            </section>

            {/* Data Sharing */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Data Sharing and Disclosure</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                We do not sell your personal data. We may share your information in the following circumstances:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                <li><strong>With your consent:</strong> When you explicitly agree to share your data</li>
                <li><strong>Service providers:</strong> With trusted third parties who assist in operating our services</li>
                <li><strong>Legal requirements:</strong> When required by law or to protect our rights</li>
                <li><strong>Business transfers:</strong> In connection with a merger, acquisition, or sale of assets</li>
                <li><strong>Public content:</strong> Content you choose to make public through our sharing features</li>
              </ul>
            </section>

            {/* Your Rights */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Your Rights</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                You have the following rights regarding your personal data:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                <li><strong>Access:</strong> Request a copy of your personal data</li>
                <li><strong>Correction:</strong> Request correction of inaccurate data</li>
                <li><strong>Deletion:</strong> Request deletion of your personal data</li>
                <li><strong>Export:</strong> Export your data in a portable format</li>
                <li><strong>Objection:</strong> Object to certain processing of your data</li>
                <li><strong>Restriction:</strong> Request restriction of processing your data</li>
                <li><strong>Withdrawal:</strong> Withdraw consent at any time</li>
              </ul>
              <p className="text-muted-foreground leading-relaxed mt-4">
                To exercise any of these rights, please contact us at privacy@atlas.com
              </p>
            </section>

            {/* Data Retention */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Data Retention</h2>
              <p className="text-muted-foreground leading-relaxed">
                We retain your personal data only for as long as necessary to fulfill the purposes outlined 
                in this privacy policy, unless a longer retention period is required or permitted by law. 
                When you delete your account, we will delete or anonymize your personal data within 30 days, 
                except where we need to retain certain information for legal or regulatory purposes.
              </p>
            </section>

            {/* Cookies */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Cookies and Tracking</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                We use cookies and similar tracking technologies to track activity on our service and store 
                certain information. You can instruct your browser to refuse all cookies or to indicate when 
                a cookie is being sent. However, if you do not accept cookies, you may not be able to use 
                some portions of our service.
              </p>
            </section>

            {/* Children's Privacy */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Children's Privacy</h2>
              <p className="text-muted-foreground leading-relaxed">
                Our service is not directed to individuals under the age of 13. We do not knowingly collect 
                personal information from children under 13. If you are a parent or guardian and you are aware 
                that your child has provided us with personal data, please contact us so we can take appropriate action.
              </p>
            </section>

            {/* International Data Transfers */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">International Data Transfers</h2>
              <p className="text-muted-foreground leading-relaxed">
                Your information may be transferred to and maintained on computers located outside of your state, 
                province, country, or other governmental jurisdiction where data protection laws may differ. 
                We ensure that appropriate safeguards are in place to protect your data in accordance with this 
                privacy policy and applicable laws.
              </p>
            </section>

            {/* Changes to Policy */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Changes to This Privacy Policy</h2>
              <p className="text-muted-foreground leading-relaxed">
                We may update our privacy policy from time to time. We will notify you of any changes by posting 
                the new privacy policy on this page and updating the "Last updated" date. You are advised to review 
                this privacy policy periodically for any changes. Changes to this privacy policy are effective when 
                they are posted on this page.
              </p>
            </section>

            {/* Contact */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Contact Us</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                If you have any questions about this privacy policy or our practices, please contact us at:
              </p>
              <div className="bg-muted/50 p-6 rounded-lg">
                <p className="text-muted-foreground">
                  <strong>Email:</strong> privacy@atlas.com<br />
                  <strong>Website:</strong> <Link to="/" className="text-primary hover:underline">atlas.com</Link>
                </p>
              </div>
            </section>
          </div>
        </div>

        {/* Back to Home Button */}
        <div className="mt-12 text-center">
          <Link to="/">
            <Button size="lg" className="gap-2">
              <ArrowLeft className="h-4 w-4" />
              Back to Home
            </Button>
          </Link>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-border/40 py-8 mt-12">
        <div className="container mx-auto px-4">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="text-sm text-muted-foreground">
              © 2025 Atlas. All rights reserved.
            </p>
            <div className="flex items-center gap-4 text-sm">
              <Link to="/terms" className="text-muted-foreground hover:text-primary transition-colors">
                Terms of Service
              </Link>
              <span className="text-muted-foreground">•</span>
              <Link to="/" className="text-muted-foreground hover:text-primary transition-colors">
                Home
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}

