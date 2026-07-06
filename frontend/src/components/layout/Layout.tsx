import { NavLink, useLocation, Outlet } from "react-router-dom";
import { HiHome, HiUserGroup, HiCube, HiPuzzlePiece, HiCog } from "react-icons/hi2";
import Aurora from "../aurora/Aurora";

const navItems = [
  { to: "/", icon: HiHome, label: "首页" },
  { to: "/room", icon: HiUserGroup, label: "房间" },
  { to: "/mods", icon: HiCube, label: "MOD" },
  { to: "/friends", icon: HiUserGroup, label: "好友" },
  { to: "/plugins", icon: HiPuzzlePiece, label: "插件" },
  { to: "/settings", icon: HiCog, label: "设置" },
];

export default function Layout() {
  const location = useLocation();
  return (
    <div className="flex h-screen overflow-hidden bg-gray-50 dark:bg-gray-950 relative">
      {/* Aurora 极光背景 */}
      <Aurora
        colorStops={["#10b981", "#06b6d4", "#6366f1"]}
        blend={0.4}
        amplitude={0.8}
        speed={0.3}
      />

      {/* 侧边导航 */}
      <nav className="w-20 flex flex-col items-center py-6 gap-2
                      bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm
                      border-r border-gray-100 dark:border-gray-800 z-10">
        {/* Logo */}
        <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-emerald-400 to-cyan-500
                        flex items-center justify-center mb-4 shadow-lg shadow-emerald-500/25">
          <span className="text-white font-bold text-sm tracking-tight">EC</span>
        </div>

        {navItems.map(({ to, icon: Icon, label }) => {
          const active =
            to === "/" ? location.pathname === "/" : location.pathname.startsWith(to);
          return (
            <NavLink
              key={to}
              to={to}
              className={`w-12 h-12 rounded-xl flex flex-col items-center justify-center
                          transition-all duration-200 relative
                          ${active
                            ? "bg-emerald-50/80 dark:bg-emerald-950/80 text-emerald-600 dark:text-emerald-400"
                            : "text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-50/50 dark:hover:bg-gray-800/50"
                          }`}
            >
              <Icon className="w-5 h-5" />
              <span className="text-[10px] mt-0.5 font-medium">{label}</span>
            </NavLink>
          );
        })}
      </nav>

      {/* 主内容 */}
      <main className="flex-1 overflow-y-auto relative z-10">
        <Outlet />
      </main>
    </div>
  );
}
