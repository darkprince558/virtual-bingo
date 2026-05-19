import type { Metadata } from 'next';
import { Nunito } from 'next/font/google';
import './globals.css';
import { SettingsProvider } from '@/contexts/SettingsContext';

const nunito = Nunito({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700', '800', '900'],
  display: 'swap',
  variable: '--font-nunito',
});

export const metadata: Metadata = {
  title: 'Virtual Bingo — Team Game',
  description: 'A live, interactive corporate bingo game. Centralized cards, real-time word calls, claim review, and winner tracking.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={nunito.variable}>
      <body suppressHydrationWarning style={{ fontFamily: "'Nunito', ui-rounded, system-ui, sans-serif" }}>
        <SettingsProvider>
          {children}
        </SettingsProvider>
      </body>
    </html>
  );
}
