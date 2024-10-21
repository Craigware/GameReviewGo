import type { Metadata } from "next";
import "./globals.css";
import Navbar from "./components/Navbar";
import Foot from "./components/Foot";
import { createContext } from "react";

export const metadata: Metadata = {
  title: "Gamedir",
  description: "A social media about games.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <Navbar />
        <main>
        {children}
        </main>
      </body>
      <Foot />
    </html>
  );
}
