import { useEffect, useState } from 'react';
import type { TweetResponse } from '../types/tweet';
import Layout from '../components/Layout';
import TweetList from '../components/TweetList';
import { useSearchParams } from 'react-router-dom';

function Home() {
    const [searchParams] = useSearchParams();
    const [tweets, setTweets] = useState<TweetResponse[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        setLoading(true);
        const searchStr = searchParams.toString();
        fetch(`/api/tweets/latest${searchStr ? '?' + searchStr : ''}`)
            .then((res) => res.json())
            .then((data) => {
                setTweets(data);
                setLoading(false);
            })
            .catch((err) => {
                console.error('Error fetching tweets:', err);
                setLoading(false);
            });
    }, [searchParams]);

    return (
        <Layout>
            {loading ? (
                <p>読み込み中...</p>
            ) : (
                <>
                    <TweetList tweets={tweets} />
                </>
            )}
        </Layout>
    );
}

export default Home;