// components/TweetList.tsx
import {formatDate} from "../utils/date.ts";
import TweetCard from "../TweetCard.tsx";

import type { TweetResponse } from '../types/tweet'

type Props = {
    tweets: TweetResponse[];
};

const TweetList = ({ tweets }: Props) => {
    return (
        <div>
            {tweets && tweets.length ? tweets.map(d => (
                <div key={d.date}>
                    <h2 className="bg-gray-500 m-[3px] p-2 text-gray-100">
                        <a href={`/dates/${d.date}`}>{formatDate(d.date)}</a>
                    </h2>
                    <div>
                        {d.tweets.map(t => (
                            <TweetCard key={t.id} tweet={t} />
                        ))}
                    </div>
                </div>
            )) : (
                <p>データがありません</p>
            )}
        </div>
    );
};

export default TweetList;