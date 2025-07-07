import { Outlet } from 'react-router-dom';
import { Toaster } from '@/components/ui/toaster';

export default function RootLayout() {
  return (
    <div className="font-body antialiased">
      <Outlet />
      <Toaster />
    </div>
  );
}