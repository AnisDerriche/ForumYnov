import React, { useState } from "react";
import { motion } from "framer-motion";

const bands = [
  { name: "Queen", image: "queen.jpg" },
  { name: "Pink Floyd", image: "pink_floyd.jpg" },
  { name: "Scorpions", image: "scorpions.jpg" },
  { name: "AC/DC", image: "acdc.jpg" },
  { name: "Pearl Jam", image: "pearl_jam.jpg" },
  { name: "Genesis", image: "genesis.jpg" },
  { name: "Kintsugi", image: "Kintsugi.jpg" },
  { name: "Alpha Wann - UMLA", image: "alpha-wann-umla.png.webp" }
];

export default function GroupieTracker() {
  const [search, setSearch] = useState("");

  const filteredBands = bands.filter((band) =>
    band.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-orange-600 flex flex-col items-center py-10">
      <h1 className="text-4xl font-bold text-white mb-6">Groupie Tracker</h1>
      <input
        type="text"
        placeholder="Search for groups..."
        className="px-4 py-2 rounded-md border border-gray-300 w-1/2 mb-6"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />
      <div className="grid grid-cols-3 gap-6">
        {filteredBands.map((band, index) => (
          <BandCard key={index} band={band} />
        ))}
      </div>
    </div>
  );
}

function BandCard({ band }) {
  const [flipped, setFlipped] = useState(false);

  return (
    <motion.div
      className="w-40 h-40 relative cursor-pointer"
      onMouseEnter={() => setFlipped(true)}
      onMouseLeave={() => setFlipped(false)}
    >
      <motion.div
        className="absolute inset-0 w-full h-full rounded-xl shadow-lg"
        initial={false}
        animate={{ rotateY: flipped ? 180 : 0 }}
        transition={{ duration: 0.6 }}
        style={{ transformStyle: "preserve-3d" }}
      >
        {!flipped ? (
          <img src={band.image} alt={band.name} className="w-full h-full object-cover rounded-xl" />
        ) : (
          <div className="w-full h-full bg-gray-800 text-white flex items-center justify-center rounded-xl">
            <span className="text-lg font-bold">{band.name}</span>
          </div>
        )}
      </motion.div>
    </motion.div>
  );
}
