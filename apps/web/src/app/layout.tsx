import type { Metadata } from "next";
import "./globals.css";

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
    <html lang="en">
      <body className="antialiased">{children}</body>
    </html>
  );
}
