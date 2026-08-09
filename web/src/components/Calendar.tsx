import React, { type JSX, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { CalendarData } from '../types/calendar';
import { getMinMaxYearMonth } from '../utils/calendar';

type CalendarProps = {
    date?: string;
};

const parseDateStr = (date?: string) => {
    if (!date || date.length < 8) return new Date();
    const year = date.slice(0, 4);
    const month = date.slice(4, 6);
    const day = date.slice(6, 8);
    return new Date(`${year}-${month}-${day}`);
};

const Calendar: React.FC<CalendarProps> = ({ date }) => {
    const dt = parseDateStr(date);

    const [year, setYear] = useState(dt.getFullYear());
    const [month, setMonth] = useState(dt.getMonth() + 1);

    const [calendarData, setCalendarData] = useState<CalendarData>({});
    const [loading, setLoading] = useState(true);

    // props.date が更新された場合に表示年月を同期する
    useEffect(() => {
        if (date && date.length >= 6) {
            const parsedYear = parseInt(date.slice(0, 4), 10);
            const parsedMonth = parseInt(date.slice(4, 6), 10);
            if (!isNaN(parsedYear) && !isNaN(parsedMonth)) {
                setYear(parsedYear);
                setMonth(parsedMonth);
            }
        }
    }, [date]);

    useEffect(() => {
        fetch('/api/calendar')
            .then((res) => res.json())
            .then((data: CalendarData) => {
                setCalendarData(data);
                setLoading(false);

                // 日付指定がなく、データが存在する場合は最新データがある年月を初期表示にする
                if (!date) {
                    const bounds = getMinMaxYearMonth(data);
                    if (bounds) {
                        setYear(bounds.maxYear);
                        setMonth(bounds.maxMonth);
                    }
                }
            })
            .catch((err) => {
                console.error('Error fetching calendar:', err);
                setLoading(false);
            });
    }, [date]);

    const bounds = getMinMaxYearMonth(calendarData);

    const isPrevMonthDisabled = bounds
        ? (year < bounds.minYear || (year === bounds.minYear && month <= bounds.minMonth))
        : false;

    const isNextMonthDisabled = bounds
        ? (year > bounds.maxYear || (year === bounds.maxYear && month >= bounds.maxMonth))
        : false;

    const getDaysInMonth = (year: number, month: number) =>
        new Date(year, month, 0).getDate();

    const isTweetDay = (y: number, m: number, d: number): boolean =>
        calendarData?.[y]?.[m]?.includes(d);

    const getWeekday = (year: number, month: number, day: number) =>
        new Date(year, month - 1, day).getDay();

    const goToPrevMonth = () => {
        if (isPrevMonthDisabled) return;
        if (month === 1) {
            setYear(year - 1);
            setMonth(12);
        } else {
            setMonth(month - 1);
        }
    };

    const goToNextMonth = () => {
        if (isNextMonthDisabled) return;
        if (month === 12) {
            setYear(year + 1);
            setMonth(1);
        } else {
            setMonth(month + 1);
        }
    };

    if (loading) return <div>読み込み中...</div>;

    const daysInMonth = getDaysInMonth(year, month);
    const firstWeekday = getWeekday(year, month, 1);

    const rows = [];
    let cells: JSX.Element[] = [];

    for (let i = 0; i < firstWeekday; i++) {
        cells.push(<td key={`empty-${i}`}></td>);
    }

    for (let d = 1; d <= daysInMonth; d++) {
        const tweetExists = isTweetDay(year, month, d);
        const ymd = `${year}${String(month).padStart(2, '0')}${String(d).padStart(2, '0')}`;

        const cell = tweetExists ? (
            <td key={d} className="py-1 text-xs sm:text-sm text-blue-600 underline font-medium">
                <Link to={`/dates/${ymd}`}>{d}</Link>
            </td>
        ) : (
            <td key={d} className="py-1 text-xs sm:text-sm text-gray-700">{d}</td>
        );
        cells.push(cell);

        if (cells.length % 7 === 0 || d === daysInMonth) {
            rows.push(<tr key={`row-${d}`}>{cells}</tr>);
            cells = [];
        }
    }

    return (
        <div className="mt-2 sm:mt-4">
            <div className="flex justify-between items-center mb-2">
                <button
                    onClick={goToPrevMonth}
                    disabled={isPrevMonthDisabled}
                    className={`px-3 py-1 rounded text-sm transition-colors cursor-pointer ${
                        isPrevMonthDisabled
                            ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                            : 'bg-gray-200 hover:bg-gray-300'
                    }`}
                >＜</button>
                <h2 className="text-base sm:text-lg font-semibold">
                    {year}年{month}月
                </h2>
                <button
                    onClick={goToNextMonth}
                    disabled={isNextMonthDisabled}
                    className={`px-3 py-1 rounded text-sm transition-colors cursor-pointer ${
                        isNextMonthDisabled
                            ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                            : 'bg-gray-200 hover:bg-gray-300'
                    }`}
                >＞</button>
            </div>

            <table className="border border-gray-300 w-full text-center text-xs sm:text-sm">
                <thead>
                <tr className="bg-gray-100">
                    <th className="py-1 text-xs sm:text-sm font-medium">日</th>
                    <th className="py-1 text-xs sm:text-sm font-medium">月</th>
                    <th className="py-1 text-xs sm:text-sm font-medium">火</th>
                    <th className="py-1 text-xs sm:text-sm font-medium">水</th>
                    <th className="py-1 text-xs sm:text-sm font-medium">木</th>
                    <th className="py-1 text-xs sm:text-sm font-medium">金</th>
                    <th className="py-1 text-xs sm:text-sm font-medium">土</th>
                </tr>
                </thead>
                <tbody>{rows}</tbody>
            </table>
        </div>
    );
};

export default Calendar;