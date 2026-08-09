import { Link } from 'react-router-dom';
import { formatDate } from "../utils/date";
import TweetCard from "./TweetCard";
import type { TweetResponse } from '../types/tweet';

type Props = {
    tweets: TweetResponse[];
};

const TweetList = ({ tweets }: Props) => {
    return (
        <div>
            {tweets && tweets.length ? tweets.map(d => (
                <div key={d.date}>
                    <h2 className="bg-gray-500 m-[3px] p-2 text-gray-100">
                        <Link to={`/dates/${d.date}`} className="hover:underline">{formatDate(d.date)}</Link>
                    </h2>
                    <div>
                        {d.tweets.map(t => (
                            <TweetCard key={t.id} tweet={t} />
                        ))}
                    </div>
                </div>
            )) : (
                <p className="p-4 text-gray-500 text-center">データがありません</p>
            )}
        </div>
    );
};

export default TweetList;