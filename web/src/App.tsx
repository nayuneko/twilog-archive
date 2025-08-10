import { Routes, Route } from 'react-router-dom'
import Home from './pages/Home'
import DatePage from './pages/DatePage'
import SearchPage from './pages/Search'

function App() {
    return (
        <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/dates/:date" element={<DatePage />} />
            <Route path="/search/" element={<SearchPage />} />
        </Routes>
    )
}

export default App;