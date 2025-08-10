import React from 'react';
import type { TweetResponseTweet } from './types/tweet'

type TweetProps = {
    tweet: TweetResponseTweet
};

const TweetCard: React.FC<TweetProps> = ({ tweet }) => {
    const formatText = (text: string) => {
        if (tweet.urls) {
            tweet.urls.map(u => {
                console.log('urls', tweet.id, u)
                text = text.replaceAll(
                    u.url,
                    `<a href="${u.expanded_url}" target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline">${u.display_url}</a>`
                )
            })
        }
        if (tweet.hashtags) {
            tweet.hashtags.map(t => {
                const tag = `#${t}`
                const url = `https://x.com/search?q=${ encodeURIComponent(tag)}`
                text = text.replaceAll(
                    tag,
                    `<a href="${url}" target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline">${tag}</a>`
                )
            })
        }
        text = text.replace(
            /https:\/\/t\.co\/[a-zA-Z0-9]+/g,
            (url) => `<a href="${url}" target="_blank" rel="noopener noreferrer" class="text-blue-500 hover:underline">${url}</a>`
        )
        return text.replace(/\n/g, '<br />');
    };
    const tweetUrl = `https://x.com/${tweet.screen_name}/status/${tweet.id}`

    return (
        <div className="border-b border-dashed border-black p-4 bg-white space-y-2 last:border-b-0">
            <div className="flex items-center gap-2">
                {tweet.name ? (
                    <>
                        <span className="font-bold">{tweet.name}</span>
                        <span className="text-sm text-gray-500">@{tweet.screen_name}</span>
                    </>
                ) : (
                    <span className="font-bold">@{tweet.screen_name}</span>
                )}
            </div>
            <div
                className="text-base leading-relaxed"
                dangerouslySetInnerHTML={{__html: formatText(tweet.text)}}
            />
            {tweet.media && tweet.media.length > 0 && (
                <div className="flex pt-2">
                    {tweet.media.map((url, i) => (
                        <img
                            key={i}
                            src={`${url}:thumb`}
                            alt="media"
                            className="rounded-xl max-h-30 object-cover pr-1"
                        />
                    ))}
                </div>
            )}
            <div className="flex items-center gap-2 text-sm text-gray-500">
                {tweet.retweeted && (
                    <div>🔁 retweeted at</div>
                )}
                {tweet.replied && (
                    <div>↪️ </div>
                )}
                {!tweet.retweeted && (
                    <div>created at</div>
                )}
                <div><a href={tweetUrl} target="_blank">{tweet.created}</a></div>
            </div>
        </div>
    );
};

export default TweetCard;