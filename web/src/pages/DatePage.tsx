import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import type { TweetResponse } from '../types/tweet';
import type { CalendarData } from '../types/calendar';
import { getAdjacentDates } from '../utils/calendar';
import Layout from "../components/Layout";
import TweetList from "../components/TweetList";

interface DateNavigatorProps {
    prevDate: string | null;
    nextDate: string | null;
}

const DateNavigator: React.FC<DateNavigatorProps> = ({ prevDate, nextDate }) => (
    <div className="flex justify-between items-center w-full my-4">
        {prevDate ? (
            <Link to={`/dates/${prevDate}`} className="text-blue-500 hover:underline pl-1.5">＜前日</Link>
        ) : (
            <span className="text-gray-400 pl-1.5 cursor-not-allowed">＜前日</span>
        )}
        {nextDate ? (
            <Link to={`/dates/${nextDate}`} className="text-blue-500 hover:underline pr-1.5">翌日＞</Link>
        ) : (
            <span className="text-gray-400 pr-1.5 cursor-not-allowed">翌日＞</span>
        )}
    </div>
);

function DatePage() {
    const { date } = useParams();
    const [tweets, setTweets] = useState<TweetResponse[]>([]);
    const [calendarData, setCalendarData] = useState<CalendarData>({});
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch('/api/calendar')
            .then((res) => res.json())
            .then((data) => setCalendarData(data))
            .catch((err) => console.error('Error fetching calendar:', err));
    }, []);

    useEffect(() => {
        if (!date) return;
        setLoading(true);
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
    }, [date]);

    const { prevDate, nextDate } = getAdjacentDates(calendarData, date || '');

    return (
        <Layout date={date}>
            {loading ? (
                <p>読み込み中...</p>
            ) : (
                <>
                    <DateNavigator prevDate={prevDate} nextDate={nextDate} />
                    <TweetList tweets={tweets}/>
                    <DateNavigator prevDate={prevDate} nextDate={nextDate} />
                </>
            )
            }
        </Layout>
    );
}

export default DatePage;