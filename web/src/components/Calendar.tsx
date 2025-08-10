import React, {type JSX, useEffect} from 'react';
import { useState } from 'react';

type CalendarData = {
    [year: string]: {
        [month: string]: number[];
    };
};

type CalendarProps = {
    date?: string;
};

const Calendar: React.FC<CalendarProps> = ({ date }) => {
    const dt = (date => {
        if (!date) return new Date()
        const year = date.slice(0, 4)
        const month = date.slice(4, 6)
        const day = date.slice(6, 8)
        return new Date(`${year}-${month}-${day}`)
    })(date)

    const [year, setYear] = useState(dt.getFullYear())
    const [month, setMonth] = useState(dt.getMonth() + 1)

    const [calendarData, setCalendarData] = useState<CalendarData>({});
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch('/api/calendar')
            .then((res) => res.json())
            .then((data) => {
                setCalendarData(data);
                setLoading(false);
            })
            .catch((err) => {
                console.error('Error fetching tweets:', err);
                setLoading(false);
            });
    }, []);

    const getDaysInMonth = (year: number, month: number) =>
        new Date(year, month, 0).getDate();

    const isTweetDay = (y: number, m: number, d: number): boolean =>
        calendarData?.[y]?.[m]?.includes(d)

    const getWeekday = (year: number, month: number, day: number) =>
        new Date(year, month - 1, day).getDay();

    const goToPrevMonth = () => {
        if (month === 1) {
            setYear(year - 1)
            setMonth(12)
        } else {
            setMonth(month - 1)
        }
    }

    const goToNextMonth = () => {
        if (month === 12) {
            setYear(year + 1)
            setMonth(1)
        } else {
            setMonth(month + 1)
        }
    }

    if (loading) return <div>読み込み中...</div>;

    const daysInMonth = getDaysInMonth(year, month)
    const firstWeekday = getWeekday(year, month, 1)

    const rows = [];
    let cells: JSX.Element[] = [];

    for (let i = 0; i < firstWeekday; i++) {
        cells.push(<td key={`empty-${i}`}></td>)
    }

    for (let d = 1; d <= daysInMonth; d++) {
        const tweetExists = isTweetDay(year, month, d)
        const ymd = `${year}${String(month).padStart(2, '0')}${String(d).padStart(2, '0')}`

        const cell = tweetExists ? (
            <td key={d} className="text-blue-600 underline">
                <a href={`/dates/${ymd}`}>{d}</a>
            </td>
        ) : (
            <td key={d}>{d}</td>
        )
        cells.push(cell)

        if (cells.length % 7 === 0 || d === daysInMonth) {
            rows.push(<tr key={`row-${d}`}>{cells}</tr>)
            cells = []
        }
    }

    return (
        <div className="mt-4">
            <div className="flex justify-between items-center mb-2">
                <button
                    onClick={goToPrevMonth}
                    className="px-3 py-1 bg-gray-200 hover:bg-gray-300 rounded"
                >＜</button>
                <h2 className="text-lg font-semibold">
                    {year}年{month}月
                </h2>
                <button
                    onClick={goToNextMonth}
                    className="px-3 py-1 bg-gray-200 hover:bg-gray-300 rounded"
                >＞</button>
            </div>

            <table className="border border-gray-300 w-full text-center">
                <thead>
                <tr className="bg-gray-100">
                    <th>日</th>
                    <th>月</th>
                    <th>火</th>
                    <th>水</th>
                    <th>木</th>
                    <th>金</th>
                    <th>土</th>
                </tr>
                </thead>
                <tbody>{rows}</tbody>
            </table>
        </div>
    );
}

export default Calendar;