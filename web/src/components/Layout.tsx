import React, { useState, useEffect } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import Calendar from './Calendar';

type Props = {
    children: React.ReactNode;
    date?: string;
    query?: string;
    excludeQuery?: string;
    searchType?: 'and' | 'or';
};

const Layout: React.FC<Props> = ({ children, date, query = '', excludeQuery = '', searchType: initialSearchType = 'and' }) => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();

    const [searchQuery, setSearchQuery] = useState(query);
    const [excludeWord, setExcludeWord] = useState(excludeQuery);
    const [searchType, setSearchType] = useState<'and' | 'or'>(initialSearchType);
    const [isSidebarOpen, setIsSidebarOpen] = useState(false);

    const [includeNormal, setIncludeNormal] = useState<boolean>(() => {
        const val = searchParams.get('normal');
        return val === null ? true : val === '1' || val === 'true';
    });
    const [includeReply, setIncludeReply] = useState<boolean>(() => {
        const val = searchParams.get('reply');
        return val === null ? true : val === '1' || val === 'true';
    });
    const [includeRT, setIncludeRT] = useState<boolean>(() => {
        const val = searchParams.get('rt');
        return val === null ? true : val === '1' || val === 'true';
    });

    useEffect(() => {
        setSearchQuery(query);
    }, [query]);

    useEffect(() => {
        setExcludeWord(excludeQuery);
    }, [excludeQuery]);

    useEffect(() => {
        setSearchType(initialSearchType);
    }, [initialSearchType]);

    useEffect(() => {
        const valNormal = searchParams.get('normal');
        setIncludeNormal(valNormal === null ? true : valNormal === '1' || valNormal === 'true');
        const valReply = searchParams.get('reply');
        setIncludeReply(valReply === null ? true : valReply === '1' || valReply === 'true');
        const valRT = searchParams.get('rt');
        setIncludeRT(valRT === null ? true : valRT === '1' || valRT === 'true');
    }, [searchParams]);

    const updateFilterParams = (normal: boolean, reply: boolean, rt: boolean) => {
        const currentParams = new URLSearchParams(window.location.search);
        if (!normal) currentParams.set('normal', '0'); else currentParams.delete('normal');
        if (!reply) currentParams.set('reply', '0'); else currentParams.delete('reply');
        if (!rt) currentParams.set('rt', '0'); else currentParams.delete('rt');

        const searchStr = currentParams.toString();
        const targetPath = window.location.pathname + (searchStr ? `?${searchStr}` : '');
        navigate(targetPath);
    };

    const handleSearchSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (searchQuery.trim() || excludeWord.trim()) {
            const params = new URLSearchParams();
            if (searchQuery.trim()) params.set('q', searchQuery.trim());
            if (excludeWord.trim()) params.set('not', excludeWord.trim());
            params.set('type', searchType);
            if (!includeNormal) params.set('normal', '0');
            if (!includeReply) params.set('reply', '0');
            if (!includeRT) params.set('rt', '0');
            navigate(`/search?${params.toString()}`);
            setIsSidebarOpen(false);
        }
    };

    return (
        <>
            <header className="bg-black p-4 sticky top-0 z-20">
                <div className="mx-auto w-full max-w-[970px] flex items-center justify-between">
                    <h1 className="text-white font-bold text-xl">
                        <Link to="/" className="hover:opacity-80 transition-opacity">𝕏 Log</Link>
                    </h1>
                    <button
                        type="button"
                        onClick={() => setIsSidebarOpen(!isSidebarOpen)}
                        className="md:hidden text-white border border-gray-700 px-3 py-1 text-sm rounded hover:bg-gray-900 transition-colors flex items-center gap-1.5 cursor-pointer"
                    >
                        <span>{isSidebarOpen ? '✕ 閉じる' : '🔍 検索・設定'}</span>
                    </button>
                </div>
            </header>

            {/* モバイル表示時の検索・設定パネル（ヘッダー直下にアコーディオン展開） */}
            {isSidebarOpen && (
                <div className="md:hidden bg-gray-100 p-4 border-b border-gray-300 space-y-4 shadow-md sticky top-[57px] z-10 max-h-[85vh] overflow-y-auto">
                    <div className="rounded-sm bg-white w-full p-3 border border-gray-200/80">
                        <div className="text-xs font-bold text-gray-500 mb-2 tracking-wide uppercase">表示対象</div>
                        <div className="flex items-center space-x-1.5">
                            <button
                                type="button"
                                onClick={() => {
                                    const next = !includeNormal;
                                    setIncludeNormal(next);
                                    updateFilterParams(next, includeReply, includeRT);
                                }}
                                className={`flex-1 py-1.5 text-xs rounded border transition-all cursor-pointer text-center font-medium ${
                                    includeNormal
                                        ? 'bg-gray-900 text-white border-gray-900 shadow-xs'
                                        : 'bg-gray-50 text-gray-400 border-gray-200 hover:bg-gray-100 hover:text-gray-600'
                                }`}
                            >
                                ツイート
                            </button>
                            <button
                                type="button"
                                onClick={() => {
                                    const next = !includeReply;
                                    setIncludeReply(next);
                                    updateFilterParams(includeNormal, next, includeRT);
                                }}
                                className={`flex-1 py-1.5 text-xs rounded border transition-all cursor-pointer text-center font-medium ${
                                    includeReply
                                        ? 'bg-gray-900 text-white border-gray-900 shadow-xs'
                                        : 'bg-gray-50 text-gray-400 border-gray-200 hover:bg-gray-100 hover:text-gray-600'
                                }`}
                            >
                                Reply
                            </button>
                            <button
                                type="button"
                                onClick={() => {
                                    const next = !includeRT;
                                    setIncludeRT(next);
                                    updateFilterParams(includeNormal, includeReply, next);
                                }}
                                className={`flex-1 py-1.5 text-xs rounded border transition-all cursor-pointer text-center font-medium ${
                                    includeRT
                                        ? 'bg-gray-900 text-white border-gray-900 shadow-xs'
                                        : 'bg-gray-50 text-gray-400 border-gray-200 hover:bg-gray-100 hover:text-gray-600'
                                }`}
                            >
                                Retweet
                            </button>
                        </div>
                    </div>
                    <div className="rounded-sm bg-white w-full p-[10px]">
                        <form onSubmit={handleSearchSubmit} className="space-y-2">
                            <div>
                                <label className="text-xs text-gray-500 block mb-0.5">検索キーワード</label>
                                <input
                                    type="text"
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    placeholder="検索キーワード"
                                    className="w-full border p-[3px] rounded-sm text-sm focus:outline-none focus:ring-1 focus:ring-black"
                                />
                            </div>
                            <div>
                                <label className="text-xs text-gray-500 block mb-0.5">除外キーワード (任意)</label>
                                <input
                                    type="text"
                                    value={excludeWord}
                                    onChange={(e) => setExcludeWord(e.target.value)}
                                    placeholder="除外する単語"
                                    className="w-full border p-[3px] rounded-sm text-sm focus:outline-none focus:ring-1 focus:ring-black bg-gray-50"
                                />
                            </div>
                            <div className="pt-1 text-center text-xs text-gray-600 flex items-center justify-between">
                                <div>
                                    <label className="pr-2 cursor-pointer">
                                        <input
                                            type="radio"
                                            name="search_type"
                                            value="and"
                                            checked={searchType === 'and'}
                                            onChange={() => setSearchType('and')}
                                        /> AND
                                    </label>
                                    <label className="cursor-pointer">
                                        <input
                                            type="radio"
                                            name="search_type"
                                            value="or"
                                            checked={searchType === 'or'}
                                            onChange={() => setSearchType('or')}
                                        /> OR
                                    </label>
                                </div>
                                <button
                                    type="submit"
                                    className="px-3 py-1 bg-gray-800 hover:bg-black text-white text-xs rounded-sm transition-colors cursor-pointer font-medium"
                                >
                                    検索
                                </button>
                            </div>
                        </form>
                    </div>
                    <div className="bg-white p-[15px] rounded-sm">
                        <Calendar date={date}/>
                    </div>
                </div>
            )}

            <div className="mx-auto w-full max-w-[970px] flex flex-col md:flex-row min-h-screen">
                <main className="w-full md:w-[640px] p-3 sm:p-4 shrink-0">{children}</main>
                {/* デスクトップ専用サイドバー */}
                <aside className="hidden md:block w-[330px] bg-gray-100 p-4 space-y-4">
                    <div className="rounded-sm bg-white w-full p-3 border border-gray-200/80">
                        <div className="text-xs font-bold text-gray-500 mb-2 tracking-wide uppercase">表示対象</div>
                        <div className="flex items-center space-x-1.5">
                            <button
                                type="button"
                                onClick={() => {
                                    const next = !includeNormal;
                                    setIncludeNormal(next);
                                    updateFilterParams(next, includeReply, includeRT);
                                }}
                                className={`flex-1 py-1.5 text-xs rounded border transition-all cursor-pointer text-center font-medium ${
                                    includeNormal
                                        ? 'bg-gray-900 text-white border-gray-900 shadow-xs'
                                        : 'bg-gray-50 text-gray-400 border-gray-200 hover:bg-gray-100 hover:text-gray-600'
                                }`}
                            >
                                ツイート
                            </button>
                            <button
                                type="button"
                                onClick={() => {
                                    const next = !includeReply;
                                    setIncludeReply(next);
                                    updateFilterParams(includeNormal, next, includeRT);
                                }}
                                className={`flex-1 py-1.5 text-xs rounded border transition-all cursor-pointer text-center font-medium ${
                                    includeReply
                                        ? 'bg-gray-900 text-white border-gray-900 shadow-xs'
                                        : 'bg-gray-50 text-gray-400 border-gray-200 hover:bg-gray-100 hover:text-gray-600'
                                }`}
                            >
                                Reply
                            </button>
                            <button
                                type="button"
                                onClick={() => {
                                    const next = !includeRT;
                                    setIncludeRT(next);
                                    updateFilterParams(includeNormal, includeReply, next);
                                }}
                                className={`flex-1 py-1.5 text-xs rounded border transition-all cursor-pointer text-center font-medium ${
                                    includeRT
                                        ? 'bg-gray-900 text-white border-gray-900 shadow-xs'
                                        : 'bg-gray-50 text-gray-400 border-gray-200 hover:bg-gray-100 hover:text-gray-600'
                                }`}
                            >
                                Retweet
                            </button>
                        </div>
                    </div>
                    <div className="rounded-sm bg-white w-full p-[10px]">
                        <form onSubmit={handleSearchSubmit} className="space-y-2">
                            <div>
                                <label className="text-xs text-gray-500 block mb-0.5">検索キーワード</label>
                                <input
                                    type="text"
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    placeholder="検索キーワード"
                                    className="w-full border p-[3px] rounded-sm text-sm focus:outline-none focus:ring-1 focus:ring-black"
                                />
                            </div>
                            <div>
                                <label className="text-xs text-gray-500 block mb-0.5">除外キーワード (任意)</label>
                                <input
                                    type="text"
                                    value={excludeWord}
                                    onChange={(e) => setExcludeWord(e.target.value)}
                                    placeholder="除外する単語"
                                    className="w-full border p-[3px] rounded-sm text-sm focus:outline-none focus:ring-1 focus:ring-black bg-gray-50"
                                />
                            </div>
                            <div className="pt-1 text-center text-xs text-gray-600 flex items-center justify-between">
                                <div>
                                    <label className="pr-2 cursor-pointer">
                                        <input
                                            type="radio"
                                            name="search_type"
                                            value="and"
                                            checked={searchType === 'and'}
                                            onChange={() => setSearchType('and')}
                                        /> AND
                                    </label>
                                    <label className="cursor-pointer">
                                        <input
                                            type="radio"
                                            name="search_type"
                                            value="or"
                                            checked={searchType === 'or'}
                                            onChange={() => setSearchType('or')}
                                        /> OR
                                    </label>
                                </div>
                                <button
                                    type="submit"
                                    className="px-3 py-1 bg-gray-800 hover:bg-black text-white text-xs rounded-sm transition-colors cursor-pointer font-medium"
                                >
                                    検索
                                </button>
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