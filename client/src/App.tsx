import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Toaster } from "react-hot-toast";
import Layout from "./components/layout/Layout";
import Home from "./pages/Home";
import Room from "./pages/Room";
import Mods from "./pages/Mods";
import Friends from "./pages/Friends";
import Plugins from "./pages/Plugins";
import Settings from "./pages/Settings";

export default function App() {
  return (
    <BrowserRouter>
      <Toaster
        position="top-center"
        toastOptions={{
          duration: 3000,
          style: {
            borderRadius: "12px",
            padding: "12px 20px",
          },
        }}
      />
      <Layout>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/room" element={<Room />} />
          <Route path="/mods" element={<Mods />} />
          <Route path="/friends" element={<Friends />} />
          <Route path="/plugins" element={<Plugins />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  );
}
