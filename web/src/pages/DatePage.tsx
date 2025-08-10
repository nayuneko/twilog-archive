import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom'
import type { TweetResponse } from '../types/tweet'
import Layout from "../components/Layout.tsx";
import TweetList from "../components/TweetList.tsx";

function DatePage() {
    const { date } = useParams()
    const [tweets, setTweets] = useState<TweetResponse[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch(`/api/tweets/dates/${date}`)
            .then((res) => res.json())
            .then((data) => {
                setTweets(data);
                setLoading(false);
            })
            .catch((err) => {
                console.error('Error fetching tweets:', err);
                setLoading(false);
            });
    }, []);

    const y = parseInt(date!.slice(0, 4), 10);
    const m = parseInt(date!.slice(4, 6), 10) - 1;
    const d = parseInt(date!.slice(6, 8), 10);
    const baseDate = new Date(y, m, d);

    // 前日
    const prevDate = new Date(baseDate);
    prevDate.setDate(baseDate.getDate() - 1);

    // 翌日
    const nextDate = new Date(baseDate);
    nextDate.setDate(baseDate.getDate() + 1);

    const formatToYYYYMMDD = (dt: Date) => {
        const yyyy = dt.getFullYear();
        const mm = String(dt.getMonth() + 1).padStart(2, '0');
        const dd = String(dt.getDate()).padStart(2, '0');
        return `${yyyy}${mm}${dd}`;
    };

    const DateNavigator = () => (
        <div className="flex justify-between items-center w-full my-4">
            <a href={`/dates/${formatToYYYYMMDD(prevDate)}`} className="text-blue-500 hover:underline pl-1.5">＜前日</a>
            <a href={`/dates/${formatToYYYYMMDD(nextDate)}`} className="text-blue-500 hover:underline pr-1.5">翌日＞</a>
        </div>
    )

    return (
        <Layout date={date}>
            {loading ? (
                <p>読み込み中...</p>
            ) : (
                <>
                    <DateNavigator/>
                    <TweetList tweets={tweets}/>
                    <DateNavigator/>
                </>
            )
            }
        </Layout>
    )
}

export default DatePage;