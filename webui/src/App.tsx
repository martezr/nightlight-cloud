import DatastoreDetails from './DatastoreDetails.tsx';
import Datastores from './Datastores.tsx';
import Home from './Home.tsx';
import InstanceConsole from './InstanceConsole.tsx';
import InstanceDetails from './InstanceDetails.tsx';
import Instances from './Instances.tsx';
import Login from './Login.tsx';
import NetworkDetails from './NetworkDetails.tsx';
import Networks from './Networks.tsx';
import NetworkTopology from './NetworkTopology.tsx';
import NewInstance from './NewInstance.tsx';
import Routers from './Routers.tsx';
import Settings from './Settings.tsx';
import Sites from './Sites.tsx';
import Subnets from './Subnets.tsx';
import Switches from './Switches.tsx';
import { ProtectedRoute } from './components/ProtectedRoute.tsx';

import { Routes, Route } from 'react-router-dom';

function App() {
  return (
    <div className="App">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<ProtectedRoute><Home /></ProtectedRoute>} />
        <Route path="/instances" element={<ProtectedRoute><Instances /></ProtectedRoute>} />
        <Route path="/instances/:id" element={<ProtectedRoute><InstanceDetails /></ProtectedRoute>} />
        <Route path="/instances/:id/console" element={<ProtectedRoute><InstanceConsole /></ProtectedRoute>} />
        <Route path="/instances/createinstance" element={<ProtectedRoute><NewInstance /></ProtectedRoute>} />
        <Route path="/datastores" element={<ProtectedRoute><Datastores /></ProtectedRoute>} />
        <Route path="/datastores/:id" element={<ProtectedRoute><DatastoreDetails /></ProtectedRoute>} />
        <Route path="/networks" element={<ProtectedRoute><Networks /></ProtectedRoute>} />
        <Route path="/networks/:id" element={<ProtectedRoute><NetworkDetails /></ProtectedRoute>} />
        <Route path="/topology" element={<ProtectedRoute><NetworkTopology /></ProtectedRoute>} />
        <Route path="/sites" element={<ProtectedRoute><Sites /></ProtectedRoute>} />
        <Route path="/networking/switches" element={<ProtectedRoute><Switches /></ProtectedRoute>} />
        <Route path="/networking/subnets" element={<ProtectedRoute><Subnets /></ProtectedRoute>} />
        <Route path="/networking/routers" element={<ProtectedRoute><Routers /></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
      </Routes>
    </div>
  );
}

export default App;
