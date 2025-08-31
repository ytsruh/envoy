import { Link, useLocation } from "react-router";
import { KeyRound, Users, Settings, BarChart3 } from "lucide-react";
import { cn } from "~/lib/utils";

const navItems = [
  { href: "/dashboard", label: "Variables", icon: KeyRound },
  { href: "#", label: "Analytics", icon: BarChart3 },
  { href: "#", label: "Team Members", icon: Users },
  { href: "#", label: "Settings", icon: Settings },
];

export function MainNav({ className, ...props }: React.HTMLAttributes<HTMLElement>) {
  const location = useLocation();
  const pathname = location.pathname;

  return (
    <nav className={cn("flex flex-col space-y-1 mt-4", className)} {...props}>
      {navItems.map((item) => (
        <Link
          key={item.label}
          to={item.href}
          className={cn(
            "flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary",
            pathname === item.href && "bg-primary text-white hover:text-white"
          )}>
          <item.icon className="h-4 w-4" />
          {item.label}
        </Link>
      ))}
    </nav>
  );
}
