import React from 'react';
import Calendar from './Calendar';

type Props = {
    children: React.ReactNode;
    date?: string;
    query?: string;
};

const Layout: React.FC<Props> = ({ children, date, query }) => {
    return (
        <>
            <header className="bg-black p-4">
                <div className="mx-auto w-[970px]">
                    <h1 className="text-white"><a href="/">𝕏 Log</a></h1>
                </div>
            </header>
            <div className="mx-auto w-[970px] flex">
                <main className="w-[640px] p-4">{children}</main>
                <aside className="w-[330px] flex-1 bg-gray-100 p-4">
                    <div className="rounded-sm bg-white w-full p-[10px]">
                        <div>並び順：新→古 | <a href="#">古→新</a></div>
                        <div>
                            <input type="checkbox" checked/>通常&nbsp;
                            <input type="checkbox" checked/>Reply&nbsp;
                            <input type="checkbox" checked/>Retweet
                        </div>
                    </div>
                    <div className="rounded-sm bg-white w-full p-[10px]">
                        <form method="GET" action="/search">
                            <input type="text" name="q" defaultValue={query} placeholder="検索"
                                   className="w-full border p-[3px] rounded-sm"/>
                            <div className=" pt-1.5 text-center">
                                <span className="pr-1.5"><input type="radio" name="search_type"
                                                                checked/>&nbsp;AND検索</span>
                                <input type="radio" name="search_type"/>&nbsp;OR検索
                            </div>
                        </form>
                    </div>
                    <div className="mt-4 bg-white p-[15px]">
                        <Calendar date={date}/>
                    </div>
                </aside>
            </div>
        </>
    );
};

export default Layout;