export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="min-h-screen flex flex-col relative overflow-x-hidden"
      style={{
        backgroundColor: '#FAF8F5',
        color: '#1C1917',
        fontFamily: "'Nunito', ui-rounded, system-ui, sans-serif",
      }}
    >
      {/* Decorative blobs - GPU-composited so they never cause scroll paint */}
      <div
        aria-hidden="true"
        style={{
          position: 'absolute',
          top: '-200px',
          right: '-200px',
          width: '600px',
          height: '600px',
          borderRadius: '9999px',
          background: 'radial-gradient(circle, rgba(255,164,112,0.18) 0%, transparent 70%)',
          pointerEvents: 'none',
          transform: 'translateZ(0)',
          willChange: 'transform',
        }}
      />
      <div
        aria-hidden="true"
        style={{
          position: 'absolute',
          bottom: '-150px',
          left: '-150px',
          width: '500px',
          height: '500px',
          borderRadius: '9999px',
          background: 'radial-gradient(circle, rgba(124,92,252,0.10) 0%, transparent 70%)',
          pointerEvents: 'none',
          transform: 'translateZ(0)',
          willChange: 'transform',
        }}
      />

      {/* Content - isolated stacking context for smooth compositing */}
      <div
        className="relative flex flex-col min-h-screen"
        style={{ isolation: 'isolate' }}
      >
        {children}
      </div>
    </div>
  )
}
