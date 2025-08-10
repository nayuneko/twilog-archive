import { useEffect, useState } from 'react';
import type { TweetResponse } from '../types/tweet'
import Layout from '../components/Layout';
import TweetList from '../components/TweetList';

function Home() {
    const [tweets, setTweets] = useState<TweetResponse[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch('/api/tweets/latest')
            .then((res) =>  res.json() )
            .then((data) => {
                setTweets(data);
                setLoading(false);
            })
            .catch((err) => {
                console.error('Error fetching tweets:', err);
                setLoading(false);
            });
    }, []);

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
    )
}

export default Home;