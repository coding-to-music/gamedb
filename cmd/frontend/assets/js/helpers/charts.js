const defaultChartOptions = {
    chart: {
        type: 'spline',
        backgroundColor: 'rgba(0,0,0,0)',
        spacing: [4, 0, 0, 0],
    },
    title: {
        text: '',
    },
    subtitle: {
        text: '',
    },
    credits: {
        enabled: false,
    },
    xAxis: {
        title: {text: ''},
        type: 'datetime',
    },
    legend: {
        enabled: true,
        itemStyle: {
            color: '#28a745',
        },
        itemHiddenStyle: {
            color: '#666666',
        },
    },
    plotOptions: {
        series: {
            marker: {
                enabled: false,
            }
        }
    },
    colors: [
        '#28a745',
        '#7cb5ec',
        '#f15c80',
        '#f7a35c',
        '#8085e9',
        '#434348',
        '#e4d354',
        '#2b908f',
        '#f45b5b',
        '#91e8e1'
    ],
};

function plotline(val, text) {

    val = parseInt(val)

    if (val.toString().length >= 10) { // Timestamp
        val = val * 1000;
    }

    return [
        {
            value: val,
            color: 'red',
            width: 1,
            zIndex: 3,
            label: {
                formatter: function () {
                    return text;
                },
                style: {
                    color: darkMode ? '#e9ecef' : '#000000',
                },
            }
        }
    ];
}
