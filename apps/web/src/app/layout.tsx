import type { Metadata } from "next";
import { Fraunces, IBM_Plex_Mono, IBM_Plex_Sans } from "next/font/google";
import { ThemeProvider } from "@/components/ThemeProvider";
import "./globals.css";

const fontBody = IBM_Plex_Sans({
  subsets: ["latin"],
  variable: "--font-body",
  weight: ["400", "500", "600", "700"],
});

const fontDisplay = Fraunces({
  subsets: ["latin"],
  variable: "--font-display",
  weight: ["600", "700", "800"],
});

const fontMono = IBM_Plex_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "Ops Portal",
  description: "Observability + AIOps portal",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `(function(){var k='ops_portal_theme';var t=localStorage.getItem(k);var d=document.documentElement;if(t==='dark')d.setAttribute('data-theme','dark');else if(t==='light')d.setAttribute('data-theme','light');else if(t==='system')d.setAttribute('data-theme',window.matchMedia('(prefers-color-scheme:dark)').matches?'dark':'light');else d.setAttribute('data-theme','light');})();`,
          }}
        />
      </head>
      <body className={`${fontBody.variable} ${fontDisplay.variable} ${fontMono.variable} antialiased`}>
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
