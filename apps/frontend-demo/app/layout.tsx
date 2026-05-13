import type {Metadata} from 'next';
import './globals.css';
import { SettingsProvider } from '@/contexts/SettingsContext';

export const metadata: Metadata = {
  title: 'Virtual Bingo',
  description: 'A fun corporate bingo game.',
};

export default function RootLayout({children}: {children: React.ReactNode}) {
  return (
    <html lang="en">
      <body suppressHydrationWarning className="theme-indigo">
        <SettingsProvider>
          {children}
        </SettingsProvider>
      </body>
    </html>
  );
}
