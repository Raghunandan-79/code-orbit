import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
  display: "swap",
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
  display: "swap",
});

const baseUrl = process.env.NEXT_PUBLIC_APP_URL ?? "https://codeorbit.dev";

export const metadata: Metadata = {
  metadataBase: new URL(baseUrl),
  title: {
    default: "Code Orbit: Real-time Collaborative Code Editor",
    template: "%s · Code Orbit",
  },
  description:
    "Write and edit code together in real time. Code Orbit is a fast collaborative code editor with multi-language support, live presence, and zero setup.",
  keywords: [
    "collaborative code editor",
    "real-time coding",
    "pair programming",
    "code sharing",
    "online IDE",
    "code orbit",
  ],
  authors: [{ name: "Code Orbit" }],
  creator: "Code Orbit",
  openGraph: {
    type: "website",
    locale: "en_US",
    url: baseUrl,
    siteName: "Code Orbit",
    title: "Code Orbit: Real-time Collaborative Code Editor",
    description:
      "Write and edit code together in real time. Multi-language, live presence, zero setup.",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "Code Orbit: Real-time Collaborative Code Editor",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Code Orbit: Real-time Collaborative Code Editor",
    description: "Write and edit code together in real time.",
    images: ["/og-image.png"],
    creator: "@codeorbit",
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  icons: {
    icon: [
      { url: "/favicon.ico", sizes: "any" },
      { url: "/icon.svg", type: "image/svg+xml" },
    ],
    apple: [{ url: "/apple-touch-icon.png" }],
  },
  manifest: "/site.webmanifest",
};

export const viewport: Viewport = {
  themeColor: "#09090b",
  width: "device-width",
  initialScale: 1,
  maximumScale: 5,
  userScalable: true,
  viewportFit: "cover",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full`}
    >
      <body className="min-h-full flex flex-col antialiased">{children}</body>
    </html>
  );
}
