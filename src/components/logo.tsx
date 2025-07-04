import { Lock } from 'lucide-react';

export function Logo() {
  return (
    <div className="flex items-center gap-2">
      <Lock className="h-6 w-6 text-primary" />
      <span className="font-headline text-xl font-bold">Envizo</span>
    </div>
  );
}
