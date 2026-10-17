import type { Metadata } from "next";
import { WorkbenchHero } from "@/components/landing/WorkbenchHero";
import { FeaturesSection } from "@/components/landing/FeaturesSection";
import { HowItWorks } from "@/components/landing/HowItWorks";
import { CtaSection } from "@/components/landing/CtaSection";
import { SiteHeader } from "@/components/layout/SiteHeader";
import { SiteFooter } from "@/components/layout/SiteFooter";

export const metadata: Metadata = {
  title: "Code Orbit: Real-time Collaborative Code Editor",
  description:
    "Write and edit code together in real time. Multi-language support, live cursor presence, zero setup required.",
};

export default function HomePage() {
  return (
    <div className="flex flex-col min-h-screen">
      <SiteHeader />
      <main className="flex-1">
        <WorkbenchHero />
        <FeaturesSection />
        <HowItWorks />
        <CtaSection />
      </main>
      <SiteFooter />
    </div>
  );
}
