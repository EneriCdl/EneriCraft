import { useState, useEffect, createContext, useContext } from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Toaster } from "react-hot-toast";
import Layout from "./components/layout/Layout";
import Home from "./pages/Home";
import Room from "./pages/Room";
import Mods from "./pages/Mods";
import Friends from "./pages/Friends";
import Plugins from "./pages/Plugins";
import Settings from "./pages/Settings";

type Theme = "dark" | "light";
const ThemeCtx = createContext<{ theme: Theme; toggle: () => void }>({
  theme: "dark",
  toggle: () => {},
});
export const useTheme = () => useContext(ThemeCtx);

export default function App() {
  const [theme, setTheme] = useState<Theme>(() => {
    return (localStorage.getItem("enericraft-theme") === "light" ? "light" : "dark");
  });

  useEffect(() => {
    const root = document.documentElement;
    root.classList.toggle("dark", theme === "dark");
    localStorage.setItem("enericraft-theme", theme);
  }, [theme]);

  const toggle = () => setTheme((t) => (t === "dark" ? "light" : "dark"));

  return (
    <ThemeCtx.Provider value={{ theme, toggle }}>
      <BrowserRouter>
        <Toaster
          position="top-center"
          toastOptions={{
            duration: 3000,
            style: {
              borderRadius: "12px",
              padding: "12px 20px",
              background: theme === "dark" ? "#1f2937" : "#fff",
              color: theme === "dark" ? "#f3f4f6" : "#111",
            },
          }}
        />
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Home />} />
            <Route path="/room" element={<Room />} />
            <Route path="/mods" element={<Mods />} />
            <Route path="/friends" element={<Friends />} />
            <Route path="/plugins" element={<Plugins />} />
            <Route path="/settings" element={<Settings />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ThemeCtx.Provider>
  );
}
