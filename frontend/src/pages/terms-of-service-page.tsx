import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { BookOpen, ArrowLeft, FileText } from 'lucide-react'

export function TermsOfServicePage() {
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
            <FileText className="h-8 w-8 text-primary" />
          </div>
          <h1 className="text-4xl font-bold mb-4">Terms of Service</h1>
          <p className="text-muted-foreground text-lg">
            Last updated: {new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
          </p>
        </div>

        {/* Terms Content */}
        <div className="prose prose-slate dark:prose-invert max-w-none">
          <div className="space-y-8">
            {/* Introduction */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Agreement to Terms</h2>
              <p className="text-muted-foreground leading-relaxed">
                Welcome to Atlas! These Terms of Service ("Terms") govern your access to and use of Atlas's 
                website, applications, and services (collectively, the "Service"). By accessing or using our 
                Service, you agree to be bound by these Terms. If you do not agree to these Terms, please do 
                not use our Service.
              </p>
            </section>

            {/* Account Registration */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Account Registration</h2>
              <div className="space-y-4">
                <p className="text-muted-foreground leading-relaxed">
                  To use certain features of our Service, you must register for an account. When you register, 
                  you agree to:
                </p>
                <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                  <li>Provide accurate, current, and complete information</li>
                  <li>Maintain and promptly update your account information</li>
                  <li>Maintain the security of your account credentials</li>
                  <li>Accept responsibility for all activities under your account</li>
                  <li>Notify us immediately of any unauthorized access</li>
                </ul>
                <p className="text-muted-foreground leading-relaxed">
                  You must be at least 13 years old to use our Service. If you are under 18, you represent 
                  that you have your parent or guardian's permission to use the Service.
                </p>
              </div>
            </section>

            {/* Use of Service */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Acceptable Use</h2>
              <div className="space-y-4">
                <p className="text-muted-foreground leading-relaxed">
                  You agree to use the Service only for lawful purposes and in accordance with these Terms. 
                  You agree not to:
                </p>
                <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                  <li>Violate any applicable laws or regulations</li>
                  <li>Infringe on the rights of others, including intellectual property rights</li>
                  <li>Upload or transmit viruses, malware, or other malicious code</li>
                  <li>Attempt to gain unauthorized access to our systems or networks</li>
                  <li>Interfere with or disrupt the Service or servers</li>
                  <li>Use the Service to harass, abuse, or harm others</li>
                  <li>Impersonate any person or entity</li>
                  <li>Collect or store personal data about other users without consent</li>
                  <li>Use automated systems to access the Service without permission</li>
                  <li>Resell or commercially exploit the Service without authorization</li>
                </ul>
              </div>
            </section>

            {/* Content */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Your Content</h2>
              <div className="space-y-4">
                <div>
                  <h3 className="text-xl font-semibold mb-2">Ownership</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    You retain all ownership rights to the content you create, upload, or store in Atlas 
                    ("Your Content"). We do not claim any ownership rights to Your Content.
                  </p>
                </div>

                <div>
                  <h3 className="text-xl font-semibold mb-2">License to Us</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    By using our Service, you grant us a limited, worldwide, non-exclusive license to:
                  </p>
                  <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                    <li>Store, process, and display Your Content to provide the Service</li>
                    <li>Use Your Content to improve our AI features (with your permission)</li>
                    <li>Make backups and copies as necessary to provide the Service</li>
                    <li>Display Your Content when you choose to share it publicly</li>
                  </ul>
                  <p className="text-muted-foreground leading-relaxed mt-2">
                    This license ends when you delete Your Content or your account, except where we need 
                    to retain copies for legal or operational purposes.
                  </p>
                </div>

                <div>
                  <h3 className="text-xl font-semibold mb-2">Responsibility</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    You are solely responsible for Your Content and the consequences of sharing it. You 
                    represent and warrant that:
                  </p>
                  <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                    <li>You own or have the necessary rights to Your Content</li>
                    <li>Your Content does not violate any laws or third-party rights</li>
                    <li>Your Content does not contain malicious code or harmful content</li>
                  </ul>
                </div>
              </div>
            </section>

            {/* AI Features */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">AI-Powered Features</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                Our Service includes AI-powered features that may process your content to provide suggestions, 
                generate text, or enhance your experience. When using these features:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                <li>AI-generated content may not always be accurate or appropriate</li>
                <li>You are responsible for reviewing and approving AI-generated content</li>
                <li>We may use third-party AI providers to power these features</li>
                <li>You retain ownership of content generated using AI features</li>
                <li>AI features may be subject to usage limits or additional terms</li>
              </ul>
            </section>

            {/* Intellectual Property */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Intellectual Property Rights</h2>
              <div className="space-y-4">
                <p className="text-muted-foreground leading-relaxed">
                  The Service and its original content (excluding Your Content), features, and functionality 
                  are owned by Atlas and are protected by international copyright, trademark, patent, trade 
                  secret, and other intellectual property laws.
                </p>
                <p className="text-muted-foreground leading-relaxed">
                  Our trademarks, logos, and service marks may not be used without our prior written consent. 
                  All other trademarks, service marks, and trade names are the property of their respective owners.
                </p>
              </div>
            </section>

            {/* Subscription and Payment */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Subscription and Payment</h2>
              <div className="space-y-4">
                <div>
                  <h3 className="text-xl font-semibold mb-2">Paid Plans</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    Some features of our Service may require a paid subscription. By subscribing to a paid plan:
                  </p>
                  <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                    <li>You agree to pay the fees for the selected plan</li>
                    <li>Fees are billed in advance on a recurring basis</li>
                    <li>You authorize us to charge your payment method</li>
                    <li>Subscription automatically renews unless cancelled</li>
                  </ul>
                </div>

                <div>
                  <h3 className="text-xl font-semibold mb-2">Cancellation and Refunds</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    You may cancel your subscription at any time. Cancellations take effect at the end of 
                    the current billing period. We do not provide refunds for partial months or unused features, 
                    except as required by law or at our sole discretion.
                  </p>
                </div>

                <div>
                  <h3 className="text-xl font-semibold mb-2">Price Changes</h3>
                  <p className="text-muted-foreground leading-relaxed">
                    We may change our fees at any time. We will provide advance notice of fee changes and give 
                    you the opportunity to cancel before the new fees take effect.
                  </p>
                </div>
              </div>
            </section>

            {/* Data and Privacy */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Data and Privacy</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                Your privacy is important to us. Our use of your personal data is governed by our Privacy Policy, 
                which is incorporated into these Terms by reference. Please review our{' '}
                <Link to="/privacy" className="text-primary hover:underline">Privacy Policy</Link> to understand 
                our practices.
              </p>
            </section>

            {/* Termination */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Termination</h2>
              <div className="space-y-4">
                <p className="text-muted-foreground leading-relaxed">
                  We may terminate or suspend your account and access to the Service immediately, without prior 
                  notice or liability, for any reason, including if you breach these Terms.
                </p>
                <p className="text-muted-foreground leading-relaxed">
                  Upon termination:
                </p>
                <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                  <li>Your right to use the Service will immediately cease</li>
                  <li>You may export Your Content before termination takes effect</li>
                  <li>We may delete Your Content after a grace period</li>
                  <li>Provisions that should survive termination will remain in effect</li>
                </ul>
              </div>
            </section>

            {/* Disclaimers */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Disclaimers</h2>
              <div className="space-y-4 bg-muted/50 p-6 rounded-lg">
                <p className="text-muted-foreground leading-relaxed">
                  THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, EITHER 
                  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO WARRANTIES OF MERCHANTABILITY, FITNESS FOR 
                  A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
                </p>
                <p className="text-muted-foreground leading-relaxed">
                  We do not warrant that:
                </p>
                <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                  <li>The Service will be uninterrupted, secure, or error-free</li>
                  <li>Results obtained from using the Service will be accurate or reliable</li>
                  <li>Any errors in the Service will be corrected</li>
                  <li>The Service will meet your requirements</li>
                </ul>
              </div>
            </section>

            {/* Limitation of Liability */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Limitation of Liability</h2>
              <div className="space-y-4 bg-muted/50 p-6 rounded-lg">
                <p className="text-muted-foreground leading-relaxed">
                  TO THE MAXIMUM EXTENT PERMITTED BY LAW, IN NO EVENT SHALL ATLAS, ITS DIRECTORS, EMPLOYEES, 
                  PARTNERS, AGENTS, OR AFFILIATES BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, 
                  OR PUNITIVE DAMAGES, INCLUDING WITHOUT LIMITATION, LOSS OF PROFITS, DATA, USE, GOODWILL, OR 
                  OTHER INTANGIBLE LOSSES, RESULTING FROM:
                </p>
                <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4">
                  <li>Your use or inability to use the Service</li>
                  <li>Any unauthorized access to or use of our servers and/or any personal information stored therein</li>
                  <li>Any interruption or cessation of the Service</li>
                  <li>Any bugs, viruses, or other harmful code transmitted through the Service</li>
                  <li>Any errors or omissions in any content or for any loss or damage incurred as a result of your use of any content</li>
                </ul>
                <p className="text-muted-foreground leading-relaxed mt-4">
                  Our total liability shall not exceed the amount you paid us in the twelve (12) months prior 
                  to the event giving rise to the liability, or $100, whichever is greater.
                </p>
              </div>
            </section>

            {/* Indemnification */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Indemnification</h2>
              <p className="text-muted-foreground leading-relaxed">
                You agree to indemnify, defend, and hold harmless Atlas and its officers, directors, employees, 
                agents, and affiliates from and against any claims, liabilities, damages, losses, and expenses, 
                including reasonable attorney's fees, arising out of or in any way connected with:
              </p>
              <ul className="list-disc list-inside space-y-2 text-muted-foreground ml-4 mt-4">
                <li>Your access to or use of the Service</li>
                <li>Your violation of these Terms</li>
                <li>Your Content or any content you submit, post, or transmit through the Service</li>
                <li>Your violation of any rights of another party</li>
              </ul>
            </section>

            {/* Dispute Resolution */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Dispute Resolution</h2>
              <div className="space-y-4">
                <p className="text-muted-foreground leading-relaxed">
                  If you have any concerns or disputes about the Service, you agree to first try to resolve 
                  the dispute informally by contacting us at support@atlas.com.
                </p>
                <p className="text-muted-foreground leading-relaxed">
                  Any disputes arising out of or relating to these Terms or the Service will be governed by 
                  the laws of [Your Jurisdiction], without regard to its conflict of law provisions.
                </p>
              </div>
            </section>

            {/* Changes to Terms */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Changes to Terms</h2>
              <p className="text-muted-foreground leading-relaxed">
                We reserve the right to modify or replace these Terms at any time at our sole discretion. 
                We will provide notice of any material changes by posting the new Terms on this page and 
                updating the "Last updated" date. Your continued use of the Service after any changes 
                constitutes acceptance of the new Terms.
              </p>
            </section>

            {/* Severability */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Severability</h2>
              <p className="text-muted-foreground leading-relaxed">
                If any provision of these Terms is found to be unenforceable or invalid, that provision will 
                be limited or eliminated to the minimum extent necessary so that these Terms will otherwise 
                remain in full force and effect.
              </p>
            </section>

            {/* Entire Agreement */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Entire Agreement</h2>
              <p className="text-muted-foreground leading-relaxed">
                These Terms, together with our Privacy Policy and any other legal notices published by us on 
                the Service, constitute the entire agreement between you and Atlas concerning the Service.
              </p>
            </section>

            {/* Contact */}
            <section>
              <h2 className="text-2xl font-semibold mb-4">Contact Us</h2>
              <p className="text-muted-foreground leading-relaxed mb-4">
                If you have any questions about these Terms, please contact us:
              </p>
              <div className="bg-muted/50 p-6 rounded-lg">
                <p className="text-muted-foreground">
                  <strong>Email:</strong> legal@atlas.com<br />
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
              <Link to="/privacy" className="text-muted-foreground hover:text-primary transition-colors">
                Privacy Policy
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

