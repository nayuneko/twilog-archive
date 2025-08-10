import { useEffect, useState } from 'react';
import type { TweetResponse } from '../types/tweet'
import Layout from '../components/Layout';
import TweetList from '../components/TweetList';
import { useSearchParams } from 'react-router-dom';

function Search() {
    const [searchParams] = useSearchParams();
    const [tweets, setTweets] = useState<TweetResponse[]>([]);
    const [loading, setLoading] = useState(true);

    const query = searchParams.get('q') || '';

    useEffect(() => {
        const fetchResults = () => {
            if (!query) return;
            setLoading(true)
            fetch(`/api/tweets/search/?q=${encodeURIComponent(query)}`)
                .then((res) =>  res.json() )
                .then((data) => {
                    setTweets(data);
                    setLoading(false);
                })
                .catch((err) => {
                    console.error('Error fetching tweets:', err);
                    setLoading(false);
                });
        };
        fetchResults();
    }, [query]);

    return (
        <Layout query={query}>
            {loading ? (
                <p>読み込み中...</p>
            ) : (
                <TweetList tweets={tweets} />
            )
            }
        </Layout>
    );
}

export default Search;