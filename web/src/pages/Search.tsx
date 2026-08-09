import { useEffect, useState } from 'react';
import type { TweetResponse } from '../types/tweet'
import Layout from '../components/Layout';
import TweetList from '../components/TweetList';
import { useSearchParams } from 'react-router-dom';

function Search() {
    const [searchParams] = useSearchParams();
    const [tweets, setTweets] = useState<TweetResponse[]>([]);
    const [totalCount, setTotalCount] = useState<number | null>(null);
    const [loading, setLoading] = useState(true);

    const query = searchParams.get('q') || '';
    const type = (searchParams.get('type') as 'and' | 'or') || 'and';

    useEffect(() => {
        const fetchResults = () => {
            if (!query) {
                setTweets([]);
                setTotalCount(null);
                setLoading(false);
                return;
            }
            setLoading(true);
            fetch(`/api/tweets/search/?q=${encodeURIComponent(query)}&type=${encodeURIComponent(type)}`)
                .then((res) => res.json())
                .then((data) => {
                    if (data && Array.isArray(data.tweets)) {
                        setTweets(data.tweets);
                        setTotalCount(data.total_count ?? 0);
                    } else if (Array.isArray(data)) {
                        setTweets(data);
                        setTotalCount(null);
                    }
                    setLoading(false);
                })
                .catch((err) => {
                    console.error('Error fetching tweets:', err);
                    setLoading(false);
                });
        };
        fetchResults();
    }, [query, type]);

    return (
        <Layout query={query} searchType={type}>
            {loading ? (
                <p>読み込み中...</p>
            ) : (
                <>
                    {query && (
                        <div className="mb-4 bg-white p-3 rounded-sm text-sm text-gray-700 flex justify-between items-center border border-gray-200">
                            <div>
                                <span>「<strong className="text-black">{query}</strong>」の検索結果 ({type.toUpperCase()}検索)</span>
                            </div>
                            {totalCount !== null && (
                                <div className="text-xs text-gray-600">
                                    該当件数: <strong className="text-sm text-blue-600 font-bold">{totalCount.toLocaleString()}</strong> 件
                                </div>
                            )}
                        </div>
                    )}
                    <TweetList tweets={tweets} />
                </>
            )}
        </Layout>
    );
}

export default Search;