import { NavLink, useLocation } from "react-router-dom";
import { HiHome, HiUserGroup, HiCube, HiPuzzle, HiCog } from "react-icons/hi2";

const navItems = [
  { to: "/", icon: HiHome, label: "首页" },
  { to: "/room", icon: HiUserGroup, label: "房间" },
  { to: "/mods", icon: HiCube, label: "MOD" },
  { to: "/friends", icon: HiUserGroup, label: "好友" },
  { to: "/plugins", icon: HiPuzzle, label: "插件" },
  { to: "/settings", icon: HiCog, label: "设置" },
];

export default function Layout({ children }: { children: React.ReactNode }) {
  const location = useLocation();

  return (
    <div className="flex h-screen overflow-hidden">
      {/* 侧边导航 */}
      <nav className="w-20 flex flex-col items-center py-6 gap-2
                      bg-white dark:bg-gray-900
                      border-r border-gray-100 dark:border-gray-800">
        {/* Logo */}
        <div className="w-12 h-12 rounded-xl bg-primary-500 flex items-center justify-center mb-4">
          <span className="text-white font-bold text-lg">MC</span>
        </div>

        {navItems.map(({ to, icon: Icon, label }) => {
          const active = to === "/" ? location.pathname === "/" : location.pathname.startsWith(to);
          return (
            <NavLink
              key={to}
              to={to}
              className={`w-12 h-12 rounded-xl flex flex-col items-center justify-center
                          transition-all duration-200 group relative
                          ${active
                            ? "bg-primary-50 dark:bg-primary-950 text-primary-600 dark:text-primary-400"
                            : "text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800"
                          }`}
            >
              <Icon className="w-5 h-5" />
              <span className="text-[10px] mt-0.5">{label}</span>
            </NavLink>
          );
        })}
      </nav>

      {/* 主内容区 */}
      <main className="flex-1 overflow-y-auto p-6">
        {children}
      </main>
    </div>
  );
}
