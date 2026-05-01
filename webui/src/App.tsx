// src/App.jsx
//import './App.css';
import DatastoreDetails from './DatastoreDetails.tsx';
import Datastores from './Datastores.tsx';
import Home from './Home.tsx'; // Example component
import InstanceConsole from './InstanceConsole.tsx';
import InstanceDetails from './InstanceDetails.tsx';
import Instances from './Instances.tsx';
import Login from './Login.tsx';
import NetworkDetails from './NetworkDetails.tsx';
import Networks from './Networks.tsx';
import NewInstance from './NewInstance.tsx';
import Settings from './Settings.tsx';

import { Routes, Route } from 'react-router-dom';

function App() {
  return (
    <div className="App">
      {/* <DoDBanner open={true} onAccept={() => {}} onDecline={() => {}} /> */}
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<Home />} />
        <Route path="/instances" element={<Instances />} />
        <Route path="/instances/:id" element={<InstanceDetails />} />
        <Route path="/instances/:id/console" element={<InstanceConsole />} />
        <Route path="/createinstance" element={<NewInstance />} />
        <Route path="/datastores" element={<Datastores />} />
        <Route path="/datastores/:id" element={<DatastoreDetails />} />
        <Route path="/networks" element={<Networks />} />
        <Route path="/networks/:id" element={<NetworkDetails />} />
        <Route path="/settings" element={<Settings />} />
        {/* Add more routes as needed */}
      </Routes>
    </div>
  );
}

export default App;