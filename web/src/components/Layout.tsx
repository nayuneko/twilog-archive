import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import Calendar from './Calendar';

type Props = {
    children: React.ReactNode;
    date?: string;
    query?: string;
    searchType?: 'and' | 'or';
};

const Layout: React.FC<Props> = ({ children, date, query = '', searchType: initialSearchType = 'and' }) => {
    const [searchQuery, setSearchQuery] = useState(query);
    const [searchType, setSearchType] = useState<'and' | 'or'>(initialSearchType);
    const navigate = useNavigate();

    const handleSearchSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (searchQuery.trim()) {
            navigate(`/search?q=${encodeURIComponent(searchQuery.trim())}&type=${searchType}`);
        }
    };

    return (
        <>
            <header className="bg-black p-4">
                <div className="mx-auto w-[970px]">
                    <h1 className="text-white font-bold text-xl">
                        <Link to="/" className="hover:opacity-80 transition-opacity">𝕏 Log</Link>
                    </h1>
                </div>
            </header>
            <div className="mx-auto w-[970px] flex min-h-screen">
                <main className="w-[640px] p-4">{children}</main>
                <aside className="w-[330px] flex-1 bg-gray-100 p-4 space-y-4">
                    <div className="rounded-sm bg-white w-full p-[10px]">
                        <div className="text-sm text-gray-700 mb-1">並び順：新→古 | <span className="text-gray-400">古→新</span></div>
                        <div className="text-sm">
                            <label className="mr-2"><input type="checkbox" defaultChecked readOnly /> 通常</label>
                            <label className="mr-2"><input type="checkbox" defaultChecked readOnly /> Reply</label>
                            <label><input type="checkbox" defaultChecked readOnly /> Retweet</label>
                        </div>
                    </div>
                    <div className="rounded-sm bg-white w-full p-[10px]">
                        <form onSubmit={handleSearchSubmit}>
                            <input
                                type="text"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                placeholder="検索"
                                className="w-full border p-[3px] rounded-sm focus:outline-none focus:ring-1 focus:ring-black"
                            />
                            <div className="pt-1.5 text-center text-xs text-gray-600">
                                <label className="pr-2 cursor-pointer">
                                    <input
                                        type="radio"
                                        name="search_type"
                                        value="and"
                                        checked={searchType === 'and'}
                                        onChange={() => setSearchType('and')}
                                    /> AND検索
                                </label>
                                <label className="cursor-pointer">
                                    <input
                                        type="radio"
                                        name="search_type"
                                        value="or"
                                        checked={searchType === 'or'}
                                        onChange={() => setSearchType('or')}
                                    /> OR検索
                                </label>
                            </div>
                        </form>
                    </div>
                    <div className="bg-white p-[15px] rounded-sm">
                        <Calendar date={date}/>
                    </div>
                </aside>
            </div>
        </>
    );
};

export default Layout;