import { BrowserRouter, Routes, Route } from 'react-router-dom';
import NavBar from './components/NavBar';
import BottomBar from './components/BottomBar';
import Home from './pages/Home';
import Prizes from './pages/Prizes';
import Start from './pages/Start';
import DrawSupporters from './pages/DrawSupporters';
import OtherResources from './pages/OtherResources';

function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-gray-50 flex flex-col">
        <NavBar />
        <main className="flex-grow">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/prizes" element={<Prizes />} />
            <Route path="/draw-supporters" element={<DrawSupporters />} />
            <Route path="/start" element={<Start />} />
            <Route path="/other-resources" element={<OtherResources />} />
          </Routes>
        </main>
        <BottomBar />
      </div>
    </BrowserRouter>
  );
}

export default App;

