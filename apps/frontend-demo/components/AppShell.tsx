export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-[#F8F9FA] text-slate-900 font-sans flex flex-col relative selection:bg-brand-100 selection:text-brand-900">
      <div 
        className="absolute inset-0 z-0 opacity-[0.28] pointer-events-none" 
        style={{ backgroundImage: 'linear-gradient(#E2E8F0 1px, transparent 1px), linear-gradient(90deg, #E2E8F0 1px, transparent 1px)', backgroundSize: '48px 48px' }}
      ></div>
      <div className="relative z-10 flex flex-col min-h-screen">
        {children}
      </div>
    </div>
  )
}
